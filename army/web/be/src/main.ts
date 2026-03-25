import { NestFactory } from '@nestjs/core';
import { NestExpressApplication } from '@nestjs/platform-express';
import { AppModule } from './app.module';
import { join } from 'path';

async function bootstrap() {
  const app = await NestFactory.create<NestExpressApplication>(AppModule);

  // CORS for development
  if (process.env.NODE_ENV !== 'production') {
    app.enableCors({ origin: 'http://localhost:5173' });
  }

  app.setGlobalPrefix('api');

  const port = process.env.PORT || 3141;
  await app.listen(port);

  // SPA fallback: serve index.html for all non-API GET requests (production only)
  if (process.env.NODE_ENV === 'production') {
    const expressApp = app.getHttpAdapter().getInstance();
    const indexPath = join(__dirname, 'public', 'index.html');
    expressApp.get('{*path}', (_req, res) => {
      res.sendFile(indexPath);
    });
  }

  console.log(`Army Web UI API listening on http://localhost:${port}`);
}
bootstrap();
