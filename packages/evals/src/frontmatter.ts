import { EvalError } from './errors.ts';

export type Scalar = string | number | boolean;
export type Value = Scalar | Scalar[];

/** Split `---` frontmatter from the body. Supports `key: value` lines with flow lists and quoted strings. */
export function parseDoc(text: string): { data: Record<string, Value>; body: string } {
  const match = /^---\n([\s\S]*?)\n?---\n?([\s\S]*)$/.exec(text);
  if (!match) return { data: {}, body: text.trim() };
  const data: Record<string, Value> = {};
  for (const line of match[1]!.split('\n')) {
    if (line.trim() === '' || line.trimStart().startsWith('#')) continue;
    const kv = /^([A-Za-z_][\w-]*):\s*(.*)$/.exec(line);
    if (!kv) throw new EvalError(`can't parse frontmatter line: ${line}`);
    data[kv[1]!] = value(kv[2]!.trim());
  }
  return { data, body: match[2]!.trim() };
}

function value(raw: string): Value {
  if (raw.startsWith('[') && raw.endsWith(']')) {
    const inner = raw.slice(1, -1).trim();
    return inner === '' ? [] : inner.split(',').map((s) => scalar(s.trim()));
  }
  return scalar(raw);
}

function scalar(raw: string): Scalar {
  const quoted = /^(["'])(.*)\1$/.exec(raw);
  if (quoted) return quoted[2]!;
  if (raw === 'true') return true;
  if (raw === 'false') return false;
  if (/^-?\d+(\.\d+)?$/.test(raw)) return Number(raw);
  return raw;
}
