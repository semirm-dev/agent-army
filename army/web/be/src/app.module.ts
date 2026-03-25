import { Module } from '@nestjs/common';
import { ServeStaticModule } from '@nestjs/serve-static';
import { join } from 'path';
import { ArmyModule } from './army/army.module';

@Module({
  imports: [
    ArmyModule,
    ...(process.env.NODE_ENV === 'production'
      ? [
          ServeStaticModule.forRoot({
            rootPath: join(__dirname, 'public'),
            exclude: ['/api/{*path}'],
          }),
        ]
      : []),
  ],
})
export class AppModule {}
