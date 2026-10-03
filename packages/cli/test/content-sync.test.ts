import { cpSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { listFiles } from '../src/files.ts';
import { ASSETS, makePackage, makeTeam, read } from './helpers.ts';

// `sync` copies the real framework content (packages/content) into a team repo unchanged.
/** A package copy with this repo's real skills and prompts (`makePackage` leaves them out). */
function packageWithSkills(): string {
  const pkg = makePackage('0.1.0');
  for (const d of ['skills', 'prompts']) cpSync(join(ASSETS, d), join(pkg, 'assets', d), { recursive: true });
  return pkg;
}

describe('sync', () => {
  it('copies every skill file into .claude/skills/ unchanged', () => {
    const team = makeTeam(packageWithSkills());
    const files = listFiles(join(ASSETS, 'skills')).filter((f) => f.includes('/'));
    expect(files.length).toBeGreaterThan(0);
    for (const f of files) {
      expect(read(join(team, '.claude/skills', f)), f).toBe(read(join(join(ASSETS, 'skills'), f)));
    }
  });

  it('copies the review prompt into .folkways/prompts/', () => {
    const team = makeTeam(packageWithSkills());
    expect(read(join(team, '.folkways/prompts/review.md'))).toBe(read(join(ASSETS, 'prompts/review.md')));
  });
});
