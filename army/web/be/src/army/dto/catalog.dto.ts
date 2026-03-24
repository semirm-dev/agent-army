export interface CatalogPlugin {
  name: string;
  marketplace: string;
  description: string;
  tags: string[];
}

export interface CatalogSkill {
  name: string;
  source: string;
  description: string;
  tags: string[];
}

export interface TechProfile {
  detect: string[];
  plugins: string[];
  skills: string[];
}

export interface Catalog {
  version: number;
  updated_at: string;
  plugins: CatalogPlugin[];
  skills: CatalogSkill[];
  tech_profiles: Record<string, TechProfile>;
}
