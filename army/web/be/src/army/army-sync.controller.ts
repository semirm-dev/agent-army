import { Controller, Post, Body, Res } from '@nestjs/common';
import { Response } from 'express';
import * as readline from 'readline';
import { ArmyService } from './army.service';

@Controller('sync')
export class ArmySyncController {
  constructor(private readonly army: ArmyService) {}

  @Post()
  async sync(@Res() res: Response, @Body() body: { destination?: string }) {
    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');
    res.flushHeaders();

    const args = ['sync'];
    if (body?.destination) {
      if (body.destination !== 'user' && body.destination !== 'project') {
        res.status(400).json({ error: 'destination must be "user" or "project"' });
        return;
      }
      args.push('--destination', body.destination);
    }

    const child = this.army.execStream(args);

    const rl = readline.createInterface({ input: child.stdout! });

    rl.on('line', (line: string) => {
      res.write(`data: ${line}\n\n`);
    });

    child.stderr?.on('data', (data: Buffer) => {
      const errMsg = JSON.stringify({
        event: 'error',
        message: data.toString().trim(),
      });
      res.write(`data: ${errMsg}\n\n`);
    });

    child.on('close', (code: number | null) => {
      res.write(
        `data: ${JSON.stringify({ event: 'exit', code: code ?? 1 })}\n\n`,
      );
      res.end();
    });

    child.on('error', (err: Error) => {
      res.write(
        `data: ${JSON.stringify({ event: 'error', message: err.message })}\n\n`,
      );
      res.end();
    });

    // Handle client disconnect
    res.on('close', () => {
      child.kill('SIGTERM');
      rl.close();
    });
  }
}
