import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import App from './App';
import { calculate } from './api';

vi.mock('./api', () => ({ calculate: vi.fn() }));
const calculateMock = vi.mocked(calculate);

async function fillBinary(a = '12', b = '3') {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText('First number'), a);
  await user.type(screen.getByLabelText('Second number'), b);
  return user;
}

describe('calculator form', () => {
  it('starts with addition and a result placeholder', () => {
    render(<App />);
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByText('—', { exact: true })).toBeInTheDocument();
  });

  it('validates empty fields without calling the backend', async () => {
    render(<App />);
    await userEvent.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(screen.getByRole('alert')).toHaveTextContent('First number must be a valid decimal number.');
    expect(calculateMock).not.toHaveBeenCalled();
  });

  it('validates the second field', async () => {
    render(<App />);
    const user = await fillBinary('4', 'wrong');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(screen.getByRole('alert')).toHaveTextContent('Second number must be a valid decimal number.');
    expect(calculateMock).not.toHaveBeenCalled();
  });

  it('submits with Enter and displays the exact backend result', async () => {
    calculateMock.mockResolvedValue('15');
    render(<App />);
    const user = await fillBinary();
    await user.keyboard('{Enter}');
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'add', a: '12', b: '3' });
    expect(await screen.findByText('12 + 3 = 15')).toBeInTheDocument();
  });

  it('switches binary operations', async () => {
    calculateMock.mockResolvedValue('4');
    render(<App />);
    const user = await fillBinary();
    await user.click(screen.getByRole('button', { name: 'Divide' }));
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'divide', a: '12', b: '3' });
    expect(await screen.findByText('12 ÷ 3 = 4')).toBeInTheDocument();
  });

  it.each([
    ['0.1', '0.2', '0.3'],
    ['10000000000000000000000000', '0.2', '1e+25'],
    ['9007199254740992', '1', '9007199254740993'],
  ])('sends decimal strings and displays the result for %s plus %s', async (a, b, result) => {
    calculateMock.mockResolvedValue(result);
    render(<App />);
    const user = await fillBinary(a, b);
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'add', a, b });
    expect(await screen.findByText(`${a} + ${b} = ${result}`)).toBeInTheDocument();
  });

  it('displays tiny power results in scientific notation', async () => {
    calculateMock.mockResolvedValue('1e-35');
    render(<App />);
    const user = await fillBinary('0.1', '35');
    await user.click(screen.getByRole('button', { name: 'Power' }));
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'power', a: '0.1', b: '35' });
    expect(await screen.findByText('0.1 xʸ 35 = 1e-35')).toBeInTheDocument();
  });

  it.each([
    ['Square root', 'sqrt', '9', '3', '√(9) = 3'],
  ] as const)('submits %s with only one operand', async (label, operation, input, result, expression) => {
    calculateMock.mockResolvedValue(result);
    render(<App />);
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: label }));
    expect(screen.queryByLabelText('Second number')).not.toBeInTheDocument();
    await user.type(screen.getByLabelText('Number'), input);
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(calculateMock).toHaveBeenCalledWith({ operation, a: input });
    expect(await screen.findByText(expression)).toBeInTheDocument();
  });

  it('calculates y percent of x using both operands', async () => {
    calculateMock.mockResolvedValue('50');
    render(<App />);
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Percentage' }));
    await user.type(screen.getByLabelText('Number (x)'), '200');
    await user.type(screen.getByLabelText('Percent (y)'), '25');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'percentage', a: '200', b: '25' });
    expect(await screen.findByText('25% of 200 = 50')).toBeInTheDocument();
  });

  it('requires a percentage rate before submitting', async () => {
    render(<App />);
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Percentage' }));
    await user.type(screen.getByLabelText('Number (x)'), '200');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(screen.getByRole('alert')).toHaveTextContent('Percent (y) must be a valid decimal number.');
    expect(calculateMock).not.toHaveBeenCalled();
  });

  it('shows backend errors and clears them when input changes', async () => {
    calculateMock.mockRejectedValue(new Error('Cannot divide by zero.'));
    render(<App />);
    const user = await fillBinary('4', '0');
    await user.click(screen.getByRole('button', { name: 'Divide' }));
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Cannot divide by zero.');
    await user.type(screen.getByLabelText('Second number'), '1');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('handles unexpected failures', async () => {
    calculateMock.mockRejectedValue('unexpected');
    render(<App />);
    const user = await fillBinary();
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
  });

  it('disables controls while waiting and then reenables them', async () => {
    let resolve!: (value: string) => void;
    calculateMock.mockReturnValue(new Promise<string>((done) => { resolve = done; }));
    render(<App />);
    const user = await fillBinary();
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(screen.getByRole('button', { name: 'Calculating…' })).toBeDisabled();
    expect(screen.getByLabelText('First number')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
    expect(screen.getByRole('status')).toHaveTextContent('Calculating…');
    await act(async () => resolve('15'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'Calculate' })).toBeEnabled());
    expect(screen.getByRole('status')).toHaveTextContent('12 + 3 = 15');
  });

  it('clears stale answers when an operand changes', async () => {
    calculateMock.mockResolvedValue('15');
    render(<App />);
    const user = await fillBinary();
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    expect(await screen.findByText('12 + 3 = 15')).toBeInTheDocument();
    await user.type(screen.getByLabelText('First number'), '0');
    expect(screen.getByText('—', { exact: true })).toBeInTheDocument();
  });

  it('resets values, operation, and result', async () => {
    calculateMock.mockResolvedValue('4');
    render(<App />);
    const user = await fillBinary();
    await user.click(screen.getByRole('button', { name: 'Divide' }));
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    await screen.findByText('12 ÷ 3 = 4');
    await user.click(screen.getByRole('button', { name: 'Reset' }));
    expect(screen.getByLabelText('First number')).toHaveValue('');
    expect(screen.getByLabelText('Second number')).toHaveValue('');
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByText('—', { exact: true })).toBeInTheDocument();
  });
});
