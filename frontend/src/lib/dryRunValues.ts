export const MAX_VALUES = 20;

/** Splits "3, 5 8" into numbers; returns null when any part is not a positive number. */
export function parseValues(text: string): number[] | null {
  const parts = text.split(/[\s,]+/).filter(Boolean);
  const values = parts.map(Number);
  return values.every((v) => Number.isFinite(v) && v > 0) && values.length <= MAX_VALUES ? values : null;
}
