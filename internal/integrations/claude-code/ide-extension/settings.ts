/**
 * Claude Code IDE Extension Settings for Kayron AI MCP
 * Configuration schema for settings.json
 */

export interface KayronMCPSettings {
  enabled: boolean;
  host: string;
  port: number;
  apiKey: string;
  cacheTtlMinutes?: number;
  logLevel?: 'debug' | 'info' | 'warn' | 'error';
  hotkeys?: Record<string, string>;
  reconnectMaxRetries?: number;
  reconnectBackoffMs?: number;
}

export const DEFAULT_SETTINGS: KayronMCPSettings = {
  enabled: true,
  host: 'localhost',
  port: 50051,
  apiKey: '',
  cacheTtlMinutes: 60,
  logLevel: 'info',
  hotkeys: {
    'place-order': 'cmd+k m',
    'close-all': 'cmd+k c',
  },
  reconnectMaxRetries: 5,
  reconnectBackoffMs: 1000,
};

export const SETTINGS_SCHEMA = {
  type: 'object',
  properties: {
    'mcp.kayron': {
      type: 'object',
      description: 'Kayron AI MCP configuration for Claude Code IDE',
      properties: {
        enabled: {
          type: 'boolean',
          description: 'Enable Kayron AI MCP integration',
          default: true,
        },
        host: {
          type: 'string',
          description: 'MCP server hostname or IP',
          default: 'localhost',
          pattern: '^[a-zA-Z0-9.-]+$',
        },
        port: {
          type: 'integer',
          description: 'MCP server port',
          default: 50051,
          minimum: 1024,
          maximum: 65535,
        },
        apiKey: {
          type: 'string',
          description:
            'API key for authentication (from env var KAYRON_API_KEY preferred)',
        },
        cacheTtlMinutes: {
          type: 'integer',
          description: 'Tool schema cache time-to-live in minutes',
          default: 60,
          minimum: 1,
        },
        logLevel: {
          type: 'string',
          enum: ['debug', 'info', 'warn', 'error'],
          description: 'Log level for Kayron operations',
          default: 'info',
        },
        hotkeys: {
          type: 'object',
          description: 'Keyboard shortcuts for trading operations',
          additionalProperties: {
            type: 'string',
            pattern: '^(cmd|ctrl|alt|shift|meta)(\\+(cmd|ctrl|alt|shift|meta|[a-z0-9]))*$',
          },
          default: {
            'place-order': 'cmd+k m',
            'close-all': 'cmd+k c',
          },
        },
        reconnectMaxRetries: {
          type: 'integer',
          description: 'Max reconnection attempts on disconnect',
          default: 5,
          minimum: 1,
          maximum: 10,
        },
        reconnectBackoffMs: {
          type: 'integer',
          description: 'Initial backoff delay for reconnection (exponential)',
          default: 1000,
          minimum: 100,
        },
      },
      required: ['host', 'port', 'apiKey'],
    },
  },
};

/**
 * Load Kayron MCP settings from IDE context
 * Priority: Environment variable > settings.json > defaults
 */
export function loadSettings(): KayronMCPSettings {
  const apiKey =
    process.env.KAYRON_API_KEY ||
    (global as any).ideSettings?.['mcp.kayron']?.apiKey ||
    '';

  const settings: KayronMCPSettings = {
    ...DEFAULT_SETTINGS,
    ...(global as any).ideSettings?.['mcp.kayron'],
    apiKey,
  };

  return settings;
}

/**
 * Validate settings against schema
 */
export function validateSettings(settings: KayronMCPSettings): string[] {
  const errors: string[] = [];

  if (!settings.host) errors.push('mcp.kayron.host is required');
  if (!settings.port || settings.port < 1024 || settings.port > 65535) {
    errors.push('mcp.kayron.port must be between 1024 and 65535');
  }
  if (!settings.apiKey) {
    errors.push(
      'mcp.kayron.apiKey is required (set via settings.json or KAYRON_API_KEY env var)'
    );
  }
  if (settings.cacheTtlMinutes && settings.cacheTtlMinutes < 1) {
    errors.push('mcp.kayron.cacheTtlMinutes must be >= 1');
  }

  return errors;
}
