import Decimal from 'decimal.js';

const decimalPlaces = 16;
const displayWidth = 20;

export function formatResult(value: string): string {
  const number = new Decimal(value);
  if (number.isZero()) return '0';

  if (number.e < displayWidth && number.e >= -6) {
    for (let places = decimalPlaces; places >= 0; places--) {
      const rounded = number.toDecimalPlaces(places, Decimal.ROUND_HALF_UP);
      const text = rounded.toFixed();
      if (text.length <= displayWidth && !rounded.isZero()) return text;
    }
  }

  let places = decimalPlaces;
  let scientific: string;
  do {
    scientific = number.toExponential(places, Decimal.ROUND_HALF_UP).replace(/\.?0+e/, 'e');
    places--;
  } while (scientific.length > displayWidth && places >= 0);
  return scientific;
}
