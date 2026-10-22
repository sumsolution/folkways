import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const REPO = resolve(import.meta.dirname, '..');
const WORKFLOWS_DIR = join(REPO, 'workflows');
const read = (path: string): string => readFileSync(path, 'utf8');

// `sync` installs each top-level `*.yml` here as `.github/workflows/folkways-<name>.yml`.
const workflows = readdirSync(WORKFLOWS_DIR).filter((f) => f.endsWith('.yml')).sort();

describe('workflows', () => {
  it('ships the knowledge-audit workflow', () => {
    expect(workflows).toContain('knowledge-audit.yml');
  });

  describe.each(workflows)('%s', (file) => {
    const text = read(join(WORKFLOWS_DIR, file));

    it('pins every action to a full commit SHA with a version comment', () => {
      const uses = [...text.matchAll(/^\s*-?\s*uses:\s*(.+)$/gm)].map((m) => m[1]?.trim());
      expect(uses.length).toBeGreaterThan(0);
      for (const use of uses) expect(use).toMatch(/^[\w.-]+\/[\w.-]+@[0-9a-f]{40} # v\d/);
    });

    it('declares its permissions', () => {
      expect(text).toMatch(/^permissions:/m);
    });

    it('runs a skill that exists', () => {
      for (const [, skill] of text.matchAll(/prompt:\s*\|?\s*\n?\s*\/([a-z0-9-]+)/g)) {
        expect(existsSync(join(REPO, 'skills', skill ?? '', 'SKILL.md')), `skill ${skill}`).toBe(true);
      }
    });
  });

  describe('knowledge-audit.yml', () => {
    const text = read(join(WORKFLOWS_DIR, 'knowledge-audit.yml'));

    it('runs monthly and on demand', () => {
      expect(text).toMatch(/schedule:\s*\n\s*- cron: "\d+ \d+ 1 \* \*"/);
      expect(text).toMatch(/^\s*workflow_dispatch:/m);
    });

    it('runs the knowledge-audit skill', () => {
      expect(text).toMatch(/prompt:[\s|]*\/knowledge-audit\b/);
    });

    it('opens a draft PR from a chore branch, never the default branch', () => {
      expect(text).toContain('chore/knowledge-audit-');
      expect(text).toContain('gh pr create --draft');
      expect(text).not.toMatch(/git push[^\n]*\b(main|master)\b/);
    });
  });
});
