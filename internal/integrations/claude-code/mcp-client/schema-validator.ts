/**
 * Schema Validator: JSON schema parsing + parameter validation
 * Validates tool inputs against schemas, provides clear error messages
 */

import { ToolDefinition, JSONSchema } from './types';

export interface ValidationResult {
  valid: boolean;
  errors: string[];
}

export class SchemaValidator {
  /**
   * Parse tool definition
   */
  parseTool(tool: ToolDefinition): ToolDefinition {
    return {
      name: tool.name,
      description: tool.description,
      inputSchema: this.normalize(tool.inputSchema),
      outputSchema: this.normalize(tool.outputSchema),
    };
  }

  /**
   * Normalize schema (add missing fields)
   */
  normalize(schema: JSONSchema): JSONSchema {
    return {
      ...schema,
      properties: schema.properties || {},
      required: schema.required || [],
    };
  }

  /**
   * Get required field names from schema
   */
  getRequiredFields(schema: JSONSchema): string[] {
    return schema.required || [];
  }

  /**
   * Get optional field names from schema
   */
  getOptionalFields(schema: JSONSchema): string[] {
    const allFields = Object.keys(schema.properties || {});
    const required = schema.required || [];
    return allFields.filter((f) => !required.includes(f));
  }

  /**
   * Get field type from schema
   */
  getFieldType(schema: JSONSchema, fieldName: string): string | undefined {
    const field = (schema.properties || {})[fieldName];
    return field?.type;
  }

  /**
   * Get enum values from schema
   */
  getEnumValues(schema: JSONSchema, fieldName: string): any[] {
    const field = (schema.properties || {})[fieldName];
    return field?.enum || [];
  }

  /**
   * Get default value from schema
   */
  getDefaultValue(schema: JSONSchema, fieldName: string): any {
    const field = (schema.properties || {})[fieldName];
    return field?.default;
  }

  /**
   * Validate parameters against schema
   */
  validate(schema: JSONSchema, params: Record<string, unknown>): ValidationResult {
    const errors: string[] = [];
    const normalized = this.normalize(schema);

    // Check required fields
    for (const required of normalized.required || []) {
      if (!(required in params)) {
        errors.push(`Missing required field: ${required}`);
      }
    }

    if (errors.length > 0) {
      return { valid: false, errors };
    }

    // Validate each provided parameter
    for (const [key, value] of Object.entries(params)) {
      const field = (normalized.properties || {})[key];

      if (!field) {
        errors.push(`Unknown field: ${key}`);
        continue;
      }

      const fieldErrors = this.validateField(field, key, value);
      errors.push(...fieldErrors);
    }

    return {
      valid: errors.length === 0,
      errors,
    };
  }

  /**
   * Validate individual field
   */
  private validateField(schema: JSONSchema, fieldName: string, value: unknown): string[] {
    const errors: string[] = [];

    // Type validation
    if (!this.isCorrectType(schema, value)) {
      errors.push(
        `Field '${fieldName}' has wrong type. Expected ${schema.type}, got ${typeof value}`
      );
      return errors;
    }

    // Enum validation
    if (schema.enum && !schema.enum.includes(value)) {
      errors.push(
        `Field '${fieldName}' must be one of: ${schema.enum.join(', ')}. Got: ${value}`
      );
      return errors;
    }

    // String length validation
    if (schema.type === 'string' && typeof value === 'string') {
      if (schema.minLength && value.length < schema.minLength) {
        errors.push(`Field '${fieldName}' must be at least ${schema.minLength} characters`);
      }
      if (schema.maxLength && value.length > schema.maxLength) {
        errors.push(`Field '${fieldName}' must be at most ${schema.maxLength} characters`);
      }
    }

    // Number range validation
    if (schema.type === 'number' && typeof value === 'number') {
      if (schema.minimum !== undefined && value < schema.minimum) {
        errors.push(`Field '${fieldName}' must be >= ${schema.minimum}. Got: ${value}`);
      }
      if (schema.maximum !== undefined && value > schema.maximum) {
        errors.push(`Field '${fieldName}' must be <= ${schema.maximum}. Got: ${value}`);
      }
    }

    // Array validation
    if (schema.type === 'array' && Array.isArray(value)) {
      if (schema.items) {
        for (let i = 0; i < value.length; i++) {
          const item = value[i];
          if (!this.isCorrectType(schema.items, item)) {
            errors.push(
              `Field '${fieldName}[${i}]' has wrong type. Expected ${schema.items.type}, got ${typeof item}`
            );
          }
        }
      }
    }

    return errors;
  }

  /**
   * Check if value matches schema type
   */
  private isCorrectType(schema: JSONSchema, value: unknown): boolean {
    if (value === null) {
      return schema.type === 'null';
    }

    switch (schema.type) {
      case 'string':
        return typeof value === 'string';
      case 'number':
        return typeof value === 'number';
      case 'integer':
        return Number.isInteger(value);
      case 'boolean':
        return typeof value === 'boolean';
      case 'array':
        return Array.isArray(value);
      case 'object':
        return typeof value === 'object' && value !== null && !Array.isArray(value);
      case 'null':
        return value === null;
      default:
        return true;
    }
  }
}
