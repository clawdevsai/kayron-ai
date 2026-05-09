/**
 * Unit Tests: Tool Schema Parsing & Validation
 * JSON schema validation, parameter type checking
 */

import { SchemaValidator } from '../../mcp-client/schema-validator';
import { ToolDefinition, JSONSchema } from '../../mcp-client/types';

describe('SchemaValidator', () => {
  let validator: SchemaValidator;

  beforeEach(() => {
    validator = new SchemaValidator();
  });

  describe('Parse Valid JSON Schema', () => {
    it('should parse valid tool definition', () => {
      const tool: ToolDefinition = {
        name: 'place-order',
        description: 'Place trading order',
        version: '1.0.0',
        inputSchema: {
          type: 'object',
          properties: {
            symbol: { type: 'string' },
            volume: { type: 'number' },
            type: { type: 'string', enum: ['BUY', 'SELL'] },
          },
          required: ['symbol', 'volume', 'type'],
        },
        outputSchema: {
          type: 'object',
          properties: {
            ticket: { type: 'number' },
            status: { type: 'string' },
          },
        },
      };

      const parsed = validator.parseTool(tool);
      expect(parsed).toBeDefined();
      expect(parsed.name).toBe('place-order');
      expect(parsed.inputSchema).toBeDefined();
      expect(parsed.outputSchema).toBeDefined();
    });

    it('should extract required fields from schema', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
          price: { type: 'number' },
        },
        required: ['symbol', 'volume'],
      };

      const required = validator.getRequiredFields(schema);
      expect(required).toEqual(['symbol', 'volume']);
    });

    it('should extract optional fields from schema', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
          price: { type: 'number' },
          comment: { type: 'string' },
        },
        required: ['symbol', 'volume'],
      };

      const optional = validator.getOptionalFields(schema);
      expect(optional).toEqual(['price', 'comment']);
    });

    it('should extract field types from schema', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
          isActive: { type: 'boolean' },
          tags: { type: 'array', items: { type: 'string' } },
        },
      };

      const symbolType = validator.getFieldType(schema, 'symbol');
      expect(symbolType).toBe('string');

      const volumeType = validator.getFieldType(schema, 'volume');
      expect(volumeType).toBe('number');

      const isActiveType = validator.getFieldType(schema, 'isActive');
      expect(isActiveType).toBe('boolean');

      const tagsType = validator.getFieldType(schema, 'tags');
      expect(tagsType).toBe('array');
    });

    it('should extract enum values from schema', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          orderType: {
            type: 'string',
            enum: ['BUY', 'SELL', 'BUY_LIMIT', 'SELL_LIMIT'],
          },
        },
      };

      const enums = validator.getEnumValues(schema, 'orderType');
      expect(enums).toEqual(['BUY', 'SELL', 'BUY_LIMIT', 'SELL_LIMIT']);
    });

    it('should extract default values from schema', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          slippage: { type: 'number', default: 0.1 },
          timeInForce: { type: 'string', default: 'GTC' },
        },
      };

      const slippage = validator.getDefaultValue(schema, 'slippage');
      expect(slippage).toBe(0.1);

      const timeInForce = validator.getDefaultValue(schema, 'timeInForce');
      expect(timeInForce).toBe('GTC');
    });
  });

  describe('Validate Input Parameters', () => {
    it('should validate valid parameters against schema', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
          type: { type: 'string', enum: ['BUY', 'SELL'] },
        },
        required: ['symbol', 'volume', 'type'],
      };

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const result = validator.validate(schema, params);
      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
    });

    it('should reject missing required field', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
          type: { type: 'string' },
        },
        required: ['symbol', 'volume', 'type'],
      };

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        // Missing 'type'
      };

      const result = validator.validate(schema, params);
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]).toContain('type');
    });

    it('should reject wrong parameter type', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
        },
        required: ['symbol', 'volume'],
      };

      const params = {
        symbol: 'EURUSD',
        volume: 'invalid-number', // Should be number
      };

      const result = validator.validate(schema, params);
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]).toContain('volume');
    });

    it('should reject invalid enum value', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          type: { type: 'string', enum: ['BUY', 'SELL'] },
        },
        required: ['type'],
      };

      const params = {
        type: 'INVALID', // Not in enum
      };

      const result = validator.validate(schema, params);
      expect(result.valid).toBe(false);
      expect(result.errors[0]).toContain('type');
      expect(result.errors[0]).toContain('BUY');
    });

    it('should allow optional fields to be omitted', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
          comment: { type: 'string' },
        },
        required: ['symbol', 'volume'],
      };

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        // 'comment' is optional and omitted
      };

      const result = validator.validate(schema, params);
      expect(result.valid).toBe(true);
    });

    it('should validate array parameters', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          tags: { type: 'array', items: { type: 'string' } },
        },
      };

      const validParams = { tags: ['urgent', 'trading'] };
      const result1 = validator.validate(schema, validParams);
      expect(result1.valid).toBe(true);

      const invalidParams = { tags: [123, 'mixed'] };
      const result2 = validator.validate(schema, invalidParams);
      expect(result2.valid).toBe(false);
    });

    it('should validate numeric ranges', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          volume: { type: 'number', minimum: 0.01, maximum: 100 },
        },
        required: ['volume'],
      };

      const validParams1 = { volume: 0.5 };
      expect(validator.validate(schema, validParams1).valid).toBe(true);

      const invalidParams1 = { volume: 0.001 }; // Below minimum
      expect(validator.validate(schema, invalidParams1).valid).toBe(false);

      const invalidParams2 = { volume: 101 }; // Above maximum
      expect(validator.validate(schema, invalidParams2).valid).toBe(false);
    });

    it('should validate string length', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          comment: { type: 'string', minLength: 3, maxLength: 100 },
        },
      };

      const validParams = { comment: 'Good trade' };
      expect(validator.validate(schema, validParams).valid).toBe(true);

      const invalidParams1 = { comment: 'a' }; // Too short
      expect(validator.validate(schema, invalidParams1).valid).toBe(false);

      const invalidParams2 = { comment: 'x'.repeat(101) }; // Too long
      expect(validator.validate(schema, invalidParams2).valid).toBe(false);
    });
  });

  describe('Error Messages', () => {
    it('should provide clear error messages', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
          volume: { type: 'number' },
        },
        required: ['symbol', 'volume'],
      };

      const params = { symbol: 'EUR' };
      const result = validator.validate(schema, params);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]).toMatch(/volume/i);
      expect(result.errors[0]).toMatch(/required/i);
    });

    it('should format error with field name and reason', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          type: { type: 'string', enum: ['BUY', 'SELL'] },
        },
        required: ['type'],
      };

      const params = { type: 'INVALID' };
      const result = validator.validate(schema, params);

      expect(result.errors[0]).toContain('type');
      expect(result.errors[0]).toContain('BUY');
      expect(result.errors[0]).toContain('SELL');
    });
  });

  describe('Schema Normalization', () => {
    it('should normalize schema with missing properties field', () => {
      const schema: JSONSchema = {
        type: 'object',
      };

      const normalized = validator.normalize(schema);
      expect(normalized.properties).toBeDefined();
      expect(typeof normalized.properties).toBe('object');
    });

    it('should normalize schema with missing required field', () => {
      const schema: JSONSchema = {
        type: 'object',
        properties: {
          symbol: { type: 'string' },
        },
      };

      const normalized = validator.normalize(schema);
      expect(Array.isArray(normalized.required)).toBe(true);
    });
  });
});
