/**
 * Tool Schema Viewer
 * Displays JSON schema in readable format (expandable, with examples)
 */

import { ToolDefinition, JSONSchema } from '../mcp-client/types';

export interface FieldDisplay {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  default?: any;
  enum?: any[];
  minLength?: number;
  maxLength?: number;
  minimum?: number;
  maximum?: number;
  example?: any;
}

export class ToolViewer {
  /**
   * Format tool schema for display
   */
  static formatTool(tool: ToolDefinition): string {
    let output = `# ${tool.name}\n\n`;
    output += `${tool.description}\n\n`;

    output += this.formatInputSchema(tool.inputSchema, tool.name);
    output += '\n';
    output += this.formatOutputSchema(tool.outputSchema, tool.name);

    return output;
  }

  /**
   * Format input schema section
   */
  private static formatInputSchema(schema: JSONSchema, toolName: string): string {
    let output = '## Input Parameters\n\n';

    const properties = schema.properties || {};
    const required = schema.required || [];

    if (Object.keys(properties).length === 0) {
      return output + 'No input parameters.\n';
    }

    // Create table
    output += '| Parameter | Type | Required | Description | Example |\n';
    output += '|-----------|------|----------|-------------|----------|\n';

    for (const [name, prop] of Object.entries(properties)) {
      const isRequired = required.includes(name) ? '✓' : '✗';
      const type = this.formatType(prop);
      const desc = this.getDescription(prop, name);
      const example = this.getExample(prop, name, toolName);

      output += `| \`${name}\` | ${type} | ${isRequired} | ${desc} | ${example} |\n`;
    }

    // Add constraints section if any
    const constraints = this.extractConstraints(schema);
    if (constraints.length > 0) {
      output += '\n### Constraints\n';
      constraints.forEach((c) => {
        output += `- **${c.field}**: ${c.description}\n`;
      });
    }

    return output;
  }

  /**
   * Format output schema section
   */
  private static formatOutputSchema(schema: JSONSchema, toolName: string): string {
    let output = '## Output\n\n';

    const properties = schema.properties || {};

    if (Object.keys(properties).length === 0) {
      return output + 'No output fields.\n';
    }

    // Create table
    output += '| Field | Type | Description | Example |\n';
    output += '|-------|------|-------------|----------|\n';

    for (const [name, prop] of Object.entries(properties)) {
      const type = this.formatType(prop);
      const desc = this.getDescription(prop, name);
      const example = this.getExample(prop, name, toolName);

      output += `| \`${name}\` | ${type} | ${desc} | ${example} |\n`;
    }

    return output;
  }

  /**
   * Format field type for display
   */
  private static formatType(prop: any): string {
    if (!prop.type) return 'any';

    let type = prop.type;

    if (prop.enum) {
      type += ` (${prop.enum.join(' | ')})`;
    }

    if (prop.type === 'array' && prop.items) {
      type = `array[${prop.items.type || 'any'}]`;
    }

    return `\`${type}\``;
  }

  /**
   * Get description from property schema
   */
  private static getDescription(prop: any, fieldName: string): string {
    if (prop.description) {
      return prop.description.substring(0, 50) + (prop.description.length > 50 ? '...' : '');
    }

    // Generate default description based on field name
    return `${fieldName} value`;
  }

  /**
   * Get example value for field
   */
  private static getExample(prop: any, fieldName: string, toolName: string): string {
    if (prop.example !== undefined) {
      return `\`${JSON.stringify(prop.example)}\``;
    }

    if (prop.default !== undefined) {
      return `\`${JSON.stringify(prop.default)}\` (default)`;
    }

    // Generate contextual example
    if (fieldName === 'symbol') return '`EURUSD`';
    if (fieldName === 'volume' || fieldName === 'lot') return '`0.1`';
    if (fieldName === 'price') return '`1.0850`';
    if (fieldName === 'type') return '`BUY`';
    if (fieldName === 'ticket' || fieldName === 'ticket_id') return '`12345`';
    if (prop.type === 'string') return '`"value"`';
    if (prop.type === 'number') return '`42`';
    if (prop.type === 'boolean') return '`true`';
    if (prop.type === 'array') return '`[...]`';

    return '—';
  }

  /**
   * Extract constraints from schema
   */
  private static extractConstraints(schema: JSONSchema): Array<{ field: string; description: string }> {
    const constraints: Array<{ field: string; description: string }> = [];
    const properties = schema.properties || {};

    for (const [name, prop] of Object.entries(properties)) {
      if ((prop as any).minimum !== undefined) {
        constraints.push({
          field: name,
          description: `Minimum value: ${(prop as any).minimum}`,
        });
      }
      if ((prop as any).maximum !== undefined) {
        constraints.push({
          field: name,
          description: `Maximum value: ${(prop as any).maximum}`,
        });
      }
      if ((prop as any).minLength !== undefined) {
        constraints.push({
          field: name,
          description: `Minimum length: ${(prop as any).minLength}`,
        });
      }
      if ((prop as any).maxLength !== undefined) {
        constraints.push({
          field: name,
          description: `Maximum length: ${(prop as any).maxLength}`,
        });
      }
      if ((prop as any).enum) {
        constraints.push({
          field: name,
          description: `Allowed values: ${(prop as any).enum.join(', ')}`,
        });
      }
    }

    return constraints;
  }

  /**
   * Expand nested properties (for complex schemas)
   */
  static expandProperties(schema: JSONSchema): FieldDisplay[] {
    const properties = schema.properties || {};
    const required = schema.required || [];

    return Object.entries(properties).map(([name, prop]) => ({
      name,
      type: (prop as any).type || 'any',
      required: required.includes(name),
      description: (prop as any).description,
      default: (prop as any).default,
      enum: (prop as any).enum,
      minLength: (prop as any).minLength,
      maxLength: (prop as any).maxLength,
      minimum: (prop as any).minimum,
      maximum: (prop as any).maximum,
      example: (prop as any).example,
    }));
  }
}
