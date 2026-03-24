import { Module } from '@nestjs/common';
import { ArmyService } from './army.service';
import { ArmyCatalogController } from './army-catalog.controller';
import { ArmyManifestController } from './army-manifest.controller';
import { ArmySyncController } from './army-sync.controller';
import { ArmyDoctorController } from './army-doctor.controller';
import { ArmyDetectController } from './army-detect.controller';
import { ArmyHealthController } from './army-health.controller';

@Module({
  providers: [ArmyService],
  controllers: [
    ArmyHealthController,
    ArmyCatalogController,
    ArmyManifestController,
    ArmySyncController,
    ArmyDoctorController,
    ArmyDetectController,
  ],
})
export class ArmyModule {}
