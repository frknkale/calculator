import { afterEach, describe, expect, it, vi } from 'vitest';
import { calculate } from './api';

afterEach(() => vi.useRealTimers());

function mockResponse(data: unknown, ok = true) {
  const fetchMock = vi.fn().mockResolvedValue({ ok, json: async () => data });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('calculate API client', () => {
  it('sends the request and returns the server result', async () => {
    const fetchMock = mockResponse({ result: 42 });
    await expect(calculate({ operation: 'add', a: 20, b: 22 })).resolves.toBe(42);
    expect(fetchMock).toHaveBeenCalledWith('/api/calculate', expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"operation":"add","a":20,"b":22}',
      signal: expect.any(AbortSignal),
    }));
  });

  it('preserves zero results', async () => {
    mockResponse({ result: 0 });
    await expect(calculate({ operation: 'sqrt', a: 0 })).resolves.toBe(0);
  });

  it('displays the backend error message', async () => {
    mockResponse({ error: { code: 'invalid_calculation', message: 'Cannot divide by zero.' } }, false);
    await expect(calculate({ operation: 'divide', a: 1, b: 0 })).rejects.toThrow('Cannot divide by zero.');
  });

  it.each([null, {}, { error: null }, { error: {} }, { error: { message: 4 } }])('handles malformed error responses: %j', async (data) => {
    mockResponse(data, false);
    await expect(calculate({ operation: 'sqrt', a: 1 })).rejects.toThrow('The calculation failed. Please try again.');
  });

  it.each([null, {}, { result: '4' }, { result: Infinity }])('rejects invalid successful response: %j', async (data) => {
    mockResponse(data);
    await expect(calculate({ operation: 'sqrt', a: 1 })).rejects.toThrow('The server returned an invalid result.');
  });

  it('handles non-JSON responses', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, json: async () => { throw new SyntaxError(); } }));
    await expect(calculate({ operation: 'sqrt', a: 1 })).rejects.toThrow('The server returned an invalid response.');
  });

  it('handles network failures', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    await expect(calculate({ operation: 'sqrt', a: 1 })).rejects.toThrow('Cannot reach the calculator service.');
  });

  it('aborts requests after ten seconds and cleans up the timer', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('fetch', vi.fn((_url, options: RequestInit) => new Promise((_resolve, reject) => {
      options.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')));
    })));
    const assertion = expect(calculate({ operation: 'sqrt', a: 1 })).rejects.toThrow('The request timed out.');
    await vi.advanceTimersByTimeAsync(10_000);
    await assertion;
    expect(vi.getTimerCount()).toBe(0);
  });
});
