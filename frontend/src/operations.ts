export const operations = [
  { value: 'add', label: 'Add', symbol: '+', unary: false, hint: 'Add two numbers together.' },
  { value: 'subtract', label: 'Subtract', symbol: '−', unary: false, hint: 'Subtract the second number from the first.' },
  { value: 'multiply', label: 'Multiply', symbol: '×', unary: false, hint: 'Multiply two numbers together.' },
  { value: 'divide', label: 'Divide', symbol: '÷', unary: false, hint: 'Divide the first number by the second.' },
  { value: 'power', label: 'Power', symbol: 'xʸ', unary: false, hint: 'Raise the first number to the power of the second.' },
  { value: 'sqrt', label: 'Square root', symbol: '√', unary: true, hint: 'Find the square root of a nonnegative number.' },
  { value: 'percentage', label: 'Percentage', symbol: '%', unary: false, hint: 'Find y percent of x: 25% of 200 = 50.' },
] as const;

export type Operation = (typeof operations)[number]['value'];

export function parseOperand(value: string, label: string): number {
  const trimmed = value.trim();
  // Accept finite decimal input, including scientific notation.
  const decimal = /^[+-]?(?:\d+\.?\d*|\.\d+)(?:e[+-]?\d+)?$/i;
  if (!decimal.test(trimmed) || !Number.isFinite(Number(trimmed))) {
    throw new Error(`${label} must be a valid, finite number.`);
  }
  return Number(trimmed);
}
