/** An expected failure: the CLI prints `folkways-eval: <message>` with no stack trace. */
export class EvalError extends Error {
  override name = 'EvalError';
}
