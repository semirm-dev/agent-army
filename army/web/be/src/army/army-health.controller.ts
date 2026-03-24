import { Controller, Get } from '@nestjs/common';

@Controller('health')
export class ArmyHealthController {
  @Get()
  health() {
    return { status: 'ok' };
  }
}
