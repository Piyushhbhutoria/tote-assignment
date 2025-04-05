export interface Plugin {
  name: string;
  description: string;
  isActive: boolean;
  config: Record<string, any>;
}

export interface PluginStats {
  eventsProcessed: number;
  lastProcessed: string | null;
  errorCount: number;
}

export interface PluginWithStats extends Plugin {
  stats: PluginStats;
} 
