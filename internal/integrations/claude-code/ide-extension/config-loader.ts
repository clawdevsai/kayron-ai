/**
 * Configuration Loader for Kayron MCP
 * Loads config from settings.json or environment variables
 */

import { KayronMCPSettings, DEFAULT_SETTINGS, validateSettings } from './settings';

export class ConfigLoader {
  /**
   * Load configuration from IDE settings + environment
   * Priority: env var > settings.json > defaults
   */
  static load(ideSettings?: Record<string, unknown>): KayronMCPSettings {
    const apiKey =
      process.env.KAYRON_API_KEY ||
      (ideSettings?.['mcp.kayron'] as any)?.apiKey ||
      '';

    const baseSettings = ideSettings?.['mcp.kayron'] || {};

    const settings: KayronMCPSettings = {
      ...DEFAULT_SETTINGS,
      ...(baseSettings as any),
      apiKey,
    };

    return settings;
  }

  /**
   * Validate loaded settings
   * Returns array of error messages (empty if valid)
   */
  static validate(settings: KayronMCPSettings): string[] {
    return validateSettings(settings);
  }

  /**
   * Load and validate in one step
   * Throws error if validation fails
   */
  static loadAndValidate(ideSettings?: Record<string, unknown>): KayronMCPSettings {
    const settings = this.load(ideSettings);
    const errors = this.validate(settings);

    if (errors.length > 0) {
      throw new Error(`Configuration validation failed: ${errors.join('; ')}`);
    }

    return settings;
  }

  /**
   * Get configuration description (for debugging)
   */
  static describe(settings: KayronMCPSettings): Record<string, unknown> {
    return {
      host: settings.host,
      port: settings.port,
      apiKeySet: !!settings.apiKey,
      cacheTtlMinutes: settings.cacheTtlMinutes,
      logLevel: settings.logLevel,
      hotkeysCount: Object.keys(settings.hotkeys || {}).length,
      reconnectMaxRetries: settings.reconnectMaxRetries,
    };
  }
}
