import { describe, expect, it } from 'vitest';
import { parseOperand } from './operations';

describe('parseOperand', () => {
  it.each([
    ['0', 0], ['-2.5', -2.5], [' .5 ', 0.5], ['+12', 12],
    ['1e3', 1000], ['-2E-2', -0.02], ['2.', 2],
  ])('accepts decimal %s', (value, expected) => {
    expect(parseOperand(value, 'Number')).toBe(expected);
  });

  it.each(['', ' ', '.', 'hello', 'Infinity', 'NaN', '1e400', '0x10', '1,000', '1 2'])('rejects %j', (value) => {
    expect(() => parseOperand(value, 'First number')).toThrow('First number must be a valid, finite number.');
  });
});
