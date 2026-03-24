import { Controller, Get } from '@nestjs/common';
import { ArmyService } from './army.service';

@Controller('doctor')
export class ArmyDoctorController {
  constructor(private readonly army: ArmyService) {}

  @Get()
  async doctor() {
    return this.army.exec(['doctor']);
  }
}
