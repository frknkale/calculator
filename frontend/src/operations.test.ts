import { describe, expect, it } from 'vitest';
import { isDecimal, parseOperand } from './operations';

describe('parseOperand', () => {
  it.each([
    ['0', '0'], ['-2.5', '-2.5'], [' .5 ', '.5'], ['+12', '+12'],
    ['1e3', '1e3'], ['-2E-2', '-2E-2'], ['2.', '2.'],
    ['0.10000000000000000001', '0.10000000000000000001'],
    ['9007199254740993', '9007199254740993'], ['1e400', '1e400'],
  ])('preserves decimal %s', (value, expected) => {
    expect(parseOperand(value, 'Number')).toBe(expected);
  });

  it.each(['', ' ', '.', 'hello', 'Infinity', 'NaN', '0x10', '1,000', '1 2'])('rejects %j', (value) => {
    expect(() => parseOperand(value, 'First number')).toThrow('First number must be a valid decimal number.');
  });

  it('rejects operands longer than 128 characters', () => {
    expect(() => parseOperand('9'.repeat(129), 'Number')).toThrow('Number must be at most 128 characters.');
  });

  it('accepts an operand of 128 characters', () => {
    const value = '9'.repeat(128);
    expect(parseOperand(value, 'Number')).toBe(value);
  });
});

describe('isDecimal', () => {
  it.each(['1\n', ' 1', '1 ', '1e', '--1'])('rejects malformed decimal strings: %j', (value) => {
    expect(isDecimal(value)).toBe(false);
  });
});
