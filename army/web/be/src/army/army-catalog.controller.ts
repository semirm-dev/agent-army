import { Controller, Get, Post } from '@nestjs/common';
import { ArmyService } from './army.service';

@Controller('catalog')
export class ArmyCatalogController {
  private cache: unknown = null;

  constructor(private readonly army: ArmyService) {}

  @Get()
  async getCatalog() {
    if (!this.cache) {
      this.cache = await this.army.exec(['catalog']);
    }
    return this.cache;
  }

  // Clear cache (useful after fetch-catalog)
  @Post('refresh')
  async refreshCatalog() {
    this.cache = null;
    this.cache = await this.army.exec(['catalog']);
    return this.cache;
  }

  @Post('fetch')
  async fetchCatalog() {
    await this.army.exec(['fetch-catalog']);
    this.cache = null;
    this.cache = await this.army.exec(['catalog']);
    return this.cache;
  }
}
