import { Injectable } from '@nestjs/common';
import { execFile, spawn, ChildProcess } from 'child_process';
import { promisify } from 'util';

const execFileAsync = promisify(execFile);

@Injectable()
export class ArmyService {
  private readonly bin: string;
  private readonly cwd: string;

  constructor() {
    this.bin = process.env.ARMY_BIN || 'army';
    this.cwd = process.env.ARMY_CWD || process.cwd();
  }

  async exec<T = unknown>(args: string[]): Promise<T> {
    const { stdout } = await execFileAsync(this.bin, [...args, '--json'], {
      cwd: this.cwd,
      env: { ...process.env },
      maxBuffer: 10 * 1024 * 1024, // 10MB
    });
    return JSON.parse(stdout) as T;
  }

  execStream(args: string[]): ChildProcess {
    return spawn(this.bin, [...args, '--json', '--yes'], {
      cwd: this.cwd,
      env: { ...process.env },
      stdio: ['ignore', 'pipe', 'pipe'],
    });
  }

  execWithInput<T = unknown>(args: string[], input: string): Promise<T> {
    return new Promise((resolve, reject) => {
      const child = spawn(this.bin, [...args, '--json'], {
        cwd: this.cwd,
        env: { ...process.env },
        stdio: ['pipe', 'pipe', 'pipe'],
      });
      let stdout = '';
      let stderr = '';
      child.stdout.on('data', (chunk: Buffer) => { stdout += chunk.toString(); });
      child.stderr.on('data', (chunk: Buffer) => { stderr += chunk.toString(); });
      child.on('close', (code) => {
        if (code !== 0) {
          reject(new Error(`army exited ${code}: ${stderr}`));
        } else {
          resolve(JSON.parse(stdout) as T);
        }
      });
      child.on('error', reject);
      child.stdin.write(input);
      child.stdin.end();
    });
  }
}
