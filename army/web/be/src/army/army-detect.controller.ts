import { Controller, Get } from '@nestjs/common';
import { ArmyService } from './army.service';

@Controller('detect')
export class ArmyDetectController {
  constructor(private readonly army: ArmyService) {}

  @Get()
  async detect() {
    return this.army.exec(['detect']);
  }
}
