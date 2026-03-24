export interface DetectResponse {
  cwd: string;
  catalog_source: string;
  manifest_path: string;
  manifest_scope: 'user' | 'project';
  manifest_exists: boolean;
  plugins_db: string;
  plugins_db_exists: boolean;
  skills_db: string;
  skills_db_exists: boolean;
  manifest_summary: {
    plugins: number;
    skills: number;
  };
}
