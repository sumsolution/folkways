// Black-box hook contract suite. Runs any implementation: FOLKWAYS_HOOK_CMD="<command>" pnpm --filter @folkways/hooks test
// The command gets the hook name appended, e.g. FOLKWAYS_HOOK_CMD="node_modules/.bin/folkways hook".
// FOLKWAYS_HOOKS=git-guard,session-end limits the run to some hooks. See README.md.
import { describe, expect, it } from 'vitest';
import { buildEnv, check, HOOKS, loadFixtures, runHook } from './harness.ts';

const command = process.env.FOLKWAYS_HOOK_CMD;
const only = process.env.FOLKWAYS_HOOKS?.split(',').filter(Boolean);
const hooks = HOOKS.filter((h) => !only || only.includes(h));

describe.skipIf(!command)('hook contract (set FOLKWAYS_HOOK_CMD to run)', () => {
  for (const hook of hooks) {
    describe(hook, () => {
      for (const [name, fixture] of loadFixtures(hook)) {
        it(`${name}: ${fixture.description}`, () => {
          const vars = buildEnv(fixture);
          const problems = fixture.steps.flatMap((step, i) => {
            const result = runHook(command!, hook, step, vars, fixture.env?.claudeProjectDir ?? true);
            return check(step.expect, result, vars).map((p) => (fixture.steps.length > 1 ? `step ${i + 1}: ${p}` : p));
          });
          expect(problems).toEqual([]);
        });
      }
    });
  }
});

describe('hook contract fixtures', () => {
  it.each(HOOKS)('%s has fixtures', (hook) => {
    expect(loadFixtures(hook).length).toBeGreaterThan(0);
  });
});
