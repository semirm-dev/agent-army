import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  // CORS for development
  if (process.env.NODE_ENV !== 'production') {
    app.enableCors({ origin: 'http://localhost:5173' });
  }

  app.setGlobalPrefix('api');

  const port = process.env.PORT || 3141;
  await app.listen(port);
  console.log(`Army Web UI API listening on http://localhost:${port}`);
}
bootstrap();
