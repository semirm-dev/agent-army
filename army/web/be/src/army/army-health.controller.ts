import { Controller, Get } from '@nestjs/common';
import { ArmyService } from './army.service';

@Controller('health')
export class ArmyHealthController {
  constructor(private readonly army: ArmyService) {}

  @Get()
  async health() {
    const version = await this.army.getVersion();
    return { status: 'ok', version };
  }
}
