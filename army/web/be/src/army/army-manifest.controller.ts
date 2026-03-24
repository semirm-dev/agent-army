import { Controller, Get, Post, Delete, Query, Param, Body } from '@nestjs/common';
import { ArmyService } from './army.service';

@Controller('manifest')
export class ArmyManifestController {
  constructor(private readonly army: ArmyService) {}

  @Get()
  async getManifest(@Query('scope') scope?: string) {
    const args = ['list'];
    if (scope === 'user' || scope === 'project') {
      args.push('--scope', scope);
    }
    return this.army.exec(args);
  }

  @Post('plugin')
  async addPlugin(@Body() body: { name: string; project?: boolean }) {
    const args = ['add', 'plugin', body.name];
    if (body.project) args.push('--project');
    return this.army.exec(args);
  }

  @Delete('plugin/:name')
  async removePlugin(@Param('name') name: string, @Query('project') project?: string) {
    const args = ['remove', 'plugin', name];
    if (project === 'true') args.push('--project');
    return this.army.exec(args);
  }

  @Post('skill')
  async addSkill(@Body() body: { name: string; project?: boolean }) {
    const args = ['add', 'skill', body.name];
    if (body.project) args.push('--project');
    return this.army.exec(args);
  }

  @Delete('skill/:name')
  async removeSkill(@Param('name') name: string, @Query('project') project?: string) {
    const args = ['remove', 'skill', name];
    if (project === 'true') args.push('--project');
    return this.army.exec(args);
  }
}
