import type { Operation } from './operations';

export interface CalculationRequest {
  operation: Operation;
  a: number;
  b?: number;
}

// Treat network data as unknown until it has been checked at runtime.
function errorMessage(data: unknown): string | undefined {
  if (typeof data !== 'object' || data === null || !('error' in data)) return;
  const error = data.error;
  if (typeof error !== 'object' || error === null || !('message' in error)) return;
  if (typeof error.message === 'string') return error.message;
}

export async function calculate(input: CalculationRequest): Promise<number> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10_000);
  try {
    const response = await fetch('/api/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
      signal: controller.signal,
    });
    let data: unknown;
    try {
      data = await response.json();
    } catch {
      throw new Error('The server returned an invalid response. Please try again.');
    }
    if (!response.ok) {
      throw new Error(errorMessage(data) ?? 'The calculation failed. Please try again.');
    }
    if (typeof data !== 'object' || data === null || !('result' in data) ||
        typeof data.result !== 'number' || !Number.isFinite(data.result)) {
      throw new Error('The server returned an invalid result. Please try again.');
    }
    return data.result;
  } catch (error) {
    if (controller.signal.aborted) {
      throw new Error('The request timed out. Please try again.');
    }
    if (error instanceof TypeError) {
      throw new Error('Cannot reach the calculator service. Please try again.');
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }
}
