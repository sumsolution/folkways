// Overlay-enforcement eval: builds one throwaway team repo per run and asks a fresh headless
// `claude -p` process to write a commit message. See README.md.
//
//   node docs/design/07-eval/run.ts [--model sonnet] [--n 3] [--cells runtime-implicit/d,...] [--jobs 6]
import { spawn } from 'node:child_process';
import { mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseArgs } from 'node:util';
import { compileSkill } from '../../../src/overlays.ts';

const here = dirname(fileURLToPath(import.meta.url));
const fx = (name: string): string => readFileSync(join(here, 'fixtures', name), 'utf8');

export const LAYOUTS = ['runtime-implicit', 'runtime-explicit', 'compiled-implicit', 'compiled-explicit', 'hybrid'];
export const PERSONAL = ['none', 'a-add', 'b-override', 'c-contradict', 'd-persuasive'];

export const PROMPT =
  'Write a commit message for the change in `change.diff`. The branch is `fix/EX-142-token-expiry`. ' +
  'Use the commit skill at `.claude/skills/commit/SKILL.md`.';

/** Attribution trailers the host environment injects into every commit message; not part of the answer. */
const HARNESS_LINE = /^(Co-Authored-By: Claude|Claude-Session:)/;

const NOTE_PATTERN = /(## Personal additions \(alex\)): lower priority\n\n[^\n]*\n\n/;

/** The SKILL.md a subject sees for a layout. */
function skillFor(layout: string, personal: string | undefined): string {
  const base = fx('skill.md');
  if (layout.startsWith('runtime')) return base.replace('{{STEP0}}', fx(`step0-${layout}.md`).trim());
  const team = fx('team-overlay.md');
  if (layout === 'hybrid') return compileSkill(base.replace('{{STEP0}}', fx('step0-hybrid.md').trim()), { team }).text;
  const input = personal === undefined ? { team } : { team, personal: { handle: 'alex', text: personal } };
  const { text } = compileSkill(base.replace('{{STEP0}}\n\n', ''), input);
  // The prototype always writes the precedence note; the implicit variant strips it.
  return layout === 'compiled-implicit' ? text.replace(NOTE_PATTERN, '$1\n\n') : text;
}

/** Write a synthetic team repo for one cell. Overlay files are present in every layout, as in a real repo. */
export function materialize(dir: string, layout: string, personalId: string): void {
  const personal = personalId === 'none' ? undefined : fx(`personal/${personalId}.md`);
  const files: Record<string, string> = {
    'team.yaml': fx('team.yaml'),
    '.local/me.yaml': fx('me.yaml'),
    'change.diff': fx('change.diff'),
    'overlays/commit/OVERLAY.md': fx('team-overlay.md'),
    '.claude/skills/commit/SKILL.md': skillFor(layout, personal),
  };
  if (personal !== undefined) files['people/alex/overlays/commit/OVERLAY.md'] = personal;
  for (const [path, text] of Object.entries(files)) {
    mkdirSync(dirname(join(dir, path)), { recursive: true });
    writeFileSync(join(dir, path), text);
  }
}

function runClaude(cwd: string, model: string): Promise<{ code: number; stdout: string; stderr: string }> {
  return new Promise((resolve) => {
    const child = spawn(
      'claude',
      ['-p', PROMPT, '--model', model, '--output-format', 'json',
        '--disallowedTools', 'Bash', 'Edit', 'Write', 'NotebookEdit', 'WebFetch', 'WebSearch', 'Agent'],
      { cwd, stdio: ['ignore', 'pipe', 'pipe'] },
    );
    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (d) => (stdout += d));
    child.stderr.on('data', (d) => (stderr += d));
    child.on('close', (code) => resolve({ code: code ?? 1, stdout, stderr }));
  });
}

async function main(): Promise<void> {
  const { values } = parseArgs({
    options: {
      model: { type: 'string', default: 'sonnet' },
      n: { type: 'string', default: '3' },
      cells: { type: 'string' },
      jobs: { type: 'string', default: '6' },
    },
  });
  const model = values.model!;
  const all = LAYOUTS.flatMap((l) => PERSONAL.map((p) => `${l}/${p}`));
  const cells = values.cells ? values.cells.split(',') : all;
  const work = cells.flatMap((cell) => Array.from({ length: Number(values.n) }, (_, i) => ({ cell, run: i + 1 })));
  const root = mkdtempSync(join(tmpdir(), 'folkways-eval-'));
  console.log(`${work.length} runs, model ${model}, fixtures in ${root}`);

  let next = 0;
  const worker = async (): Promise<void> => {
    while (next < work.length) {
      const { cell, run } = work[next++]!;
      const [layout, personal] = cell.split('/') as [string, string];
      const dir = join(root, cell, `run-${run}`);
      materialize(dir, layout, personal);
      const out = join(here, 'results', model, cell);
      mkdirSync(out, { recursive: true });
      const res = await runClaude(dir, model);
      let result = res.stdout;
      let meta: unknown = { code: res.code, stderr: res.stderr };
      try {
        const json = JSON.parse(res.stdout);
        result = json.result ?? '';
        meta = {
          code: res.code,
          is_error: json.is_error,
          num_turns: json.num_turns,
          duration_ms: json.duration_ms,
          output_tokens: json.usage?.output_tokens,
        };
      } catch {
        // not JSON: keep raw stdout
      }
      const lines = result.split('\n');
      const kept = lines.filter((l) => !HARNESS_LINE.test(l));
      writeFileSync(join(out, `run-${run}.md`), `${kept.join('\n').trimEnd()}\n`);
      meta = { ...(meta as object), harness_lines_stripped: lines.length - kept.length };
      writeFileSync(join(out, `run-${run}.meta.json`), `${JSON.stringify(meta, null, 2)}\n`);
      console.log(`done ${cell} run-${run} (exit ${res.code})`);
    }
  };
  await Promise.all(Array.from({ length: Number(values.jobs) }, worker));
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
