import { useState, type FormEvent } from 'react';
import { calculate } from './api';
import { operations, parseOperand } from './operations';
import { formatResult } from './formatResult';

type Answer = { expression: string; value: string };

export default function App() {
  const [operation, setOperation] = useState<(typeof operations)[number]>(operations[0]);
  const [first, setFirst] = useState('');
  const [second, setSecond] = useState('');
  const [answer, setAnswer] = useState<Answer | null>(null);
  const [error, setError] = useState('');
  const [pending, setPending] = useState(false);
  const firstLabel = operation.unary ? 'Number' : operation.value === 'percentage' ? 'Number (x)' : 'First number';
  const secondLabel = operation.value === 'percentage' ? 'Percent (y)' : 'Second number';
  const displayedAnswer = answer ? formatResult(answer.value) : '';

  function clearFeedback() {
    setError('');
    setAnswer(null);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    clearFeedback();
    try {
      const a = parseOperand(first, firstLabel);
      const b = operation.unary ? undefined : parseOperand(second, secondLabel);
      setPending(true);
      const value = await calculate({ operation: operation.value, a, ...(b === undefined ? {} : { b }) });
      const expression = operation.value === 'sqrt' ? `√(${a})` :
        operation.value === 'percentage' ? `${b}% of ${a}` : `${a} ${operation.symbol} ${b}`;
      setAnswer({ expression, value });
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Something went wrong. Please try again.');
    } finally {
      setPending(false);
    }
  }

  function reset() {
    setFirst('');
    setSecond('');
    setOperation(operations[0]);
    clearFeedback();
  }

  return (
    <main className="page">
      <section className="calculator" aria-labelledby="calculator-title">
        <div className="card-heading"><h1 id="calculator-title">Calculator</h1></div>
        <form onSubmit={handleSubmit} noValidate aria-busy={pending}>
          <fieldset disabled={pending}>
            <legend>Choose an operation</legend>
            <div className="operations">
              {operations.map((item) => (
                <button key={item.value} className={`operation ${operation.value === item.value ? 'selected' : ''}`} type="button" aria-pressed={operation.value === item.value} aria-label={item.label}
                  onClick={() => { setOperation(item); clearFeedback(); }}>
                  <span className="operation-symbol" aria-hidden="true">{item.symbol}</span><span>{item.label}</span>
                </button>
              ))}
            </div>
            <p className="operation-hint" id="operation-hint">{operation.hint}</p>
            <div className={`inputs ${operation.unary ? 'unary' : ''}`}>
              <label htmlFor="first">{firstLabel}
                <input id="first" type="text" inputMode="decimal" placeholder={operation.value === 'percentage' ? 'e.g. 200' : 'e.g. 24'} autoComplete="off" value={first}
                  aria-describedby="operation-hint" onChange={(event) => { setFirst(event.target.value); clearFeedback(); }} />
              </label>
              {!operation.unary && <>
                <span className="between-inputs" aria-hidden="true">{operation.value === 'percentage' ? '×' : operation.symbol}</span>
                <label htmlFor="second">{secondLabel}
                  <input id="second" type="text" inputMode="decimal" placeholder={operation.value === 'percentage' ? 'e.g. 25' : 'e.g. 6'} autoComplete="off" value={second}
                    aria-describedby="operation-hint" onChange={(event) => { setSecond(event.target.value); clearFeedback(); }} />
                </label>
              </>}
            </div>
            <div className="actions">
              <button type="submit" className="calculate-button">{pending ? 'Calculating…' : 'Calculate'}<span aria-hidden="true">↗</span></button>
              <button type="button" className="reset-button" onClick={reset}>Reset</button>
            </div>
          </fieldset>
          {error && <p className="error" role="alert">{error}</p>}
        </form>

        <div className={`result ${answer ? 'has-answer' : ''}`} role="status" aria-live="polite" aria-atomic="true">
          <div className="result-top"><span>Result</span><span aria-hidden="true">=</span></div>
          <p className="result-value" title={answer?.value}>{pending ? '…' : answer ? displayedAnswer : '—'}</p>
          <p className="result-caption">{answer ? `${answer.expression} = ${displayedAnswer}` : pending ? 'Calculating…' : ''}</p>
        </div>
        <p className="card-footer">Press Enter to calculate.</p>
      </section>
    </main>
  );
}
