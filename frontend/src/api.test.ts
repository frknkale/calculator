import { afterEach, describe, expect, it, vi } from 'vitest';
import { calculate } from './api';

afterEach(() => vi.useRealTimers());

function mockResponse(data: unknown, ok = true) {
  const fetchMock = vi.fn().mockResolvedValue({ ok, json: async () => data });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('calculate API client', () => {
  it('sends decimal strings and returns the server result', async () => {
    const fetchMock = mockResponse({ result: '0.3' });
    await expect(calculate({ operation: 'add', a: '0.1', b: '0.2' })).resolves.toBe('0.3');
    expect(fetchMock).toHaveBeenCalledWith('/api/calculate', expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"operation":"add","a":"0.1","b":"0.2"}',
      signal: expect.any(AbortSignal),
    }));
  });

  it.each(['0', '-1.25', '9007199254740993', '10000000000000000000000000.2', '0.' + '0'.repeat(34) + '1'])('preserves the result %s', async (result) => {
    mockResponse({ result });
    await expect(calculate({ operation: 'add', a: '0', b: '0' })).resolves.toBe(result);
  });

  it('omits the unused operand for square roots', async () => {
    const fetchMock = mockResponse({ result: '3' });
    await expect(calculate({ operation: 'sqrt', a: '9' })).resolves.toBe('3');
    expect(fetchMock).toHaveBeenCalledWith('/api/calculate', expect.objectContaining({
      body: '{"operation":"sqrt","a":"9"}',
    }));
  });

  it('displays the backend error message', async () => {
    mockResponse({ error: { code: 'invalid_calculation', message: 'Cannot divide by zero.' } }, false);
    await expect(calculate({ operation: 'divide', a: '1', b: '0' })).rejects.toThrow('Cannot divide by zero.');
  });

  it.each([null, {}, { error: null }, { error: {} }, { error: { message: 4 } }])('handles malformed error responses: %j', async (data) => {
    mockResponse(data, false);
    await expect(calculate({ operation: 'sqrt', a: '1' })).rejects.toThrow('The calculation failed. Please try again.');
  });

  it.each([null, {}, { result: 4 }, { result: '' }, { result: 'NaN' }, { result: 'Infinity' }, { result: '1x' }, { result: '1\n' }])('rejects invalid successful response: %j', async (data) => {
    mockResponse(data);
    await expect(calculate({ operation: 'sqrt', a: '1' })).rejects.toThrow('The server returned an invalid result.');
  });

  it('handles non-JSON responses', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, json: async () => { throw new SyntaxError(); } }));
    await expect(calculate({ operation: 'sqrt', a: '1' })).rejects.toThrow('The server returned an invalid response.');
  });

  it('handles network failures', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    await expect(calculate({ operation: 'sqrt', a: '1' })).rejects.toThrow('Cannot reach the calculator service.');
  });

  it('aborts requests after ten seconds and cleans up the timer', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('fetch', vi.fn((_url, options: RequestInit) => new Promise((_resolve, reject) => {
      options.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')));
    })));
    const assertion = expect(calculate({ operation: 'sqrt', a: '1' })).rejects.toThrow('The request timed out.');
    await vi.advanceTimersByTimeAsync(10_000);
    await assertion;
    expect(vi.getTimerCount()).toBe(0);
  });
});
