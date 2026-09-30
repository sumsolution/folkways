import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import { armPackage, buildTeamRepo, WORKTREE, workspaceSkills } from '../src/workspace.ts';

const repoRoot = resolve(import.meta.dirname, '../../..');
const scratch = () => mkdtempSync(join(tmpdir(), 'evals-ws-'));

describe('armPackage', () => {
  it('uses the working tree CLI for WORKTREE', () => {
    expect(armPackage(repoRoot, WORKTREE, scratch())).toBe(join(repoRoot, 'packages/cli'));
  });

  it('extracts the CLI package at a git ref', () => {
    const pkg = armPackage(repoRoot, 'HEAD', scratch());
    expect(existsSync(join(pkg, 'src/cli.ts'))).toBe(true);
    expect(existsSync(join(pkg, 'assets/instructions.md'))).toBe(true);
  });

  it('rejects a ref that is not a commit', () => {
    expect(() => armPackage(repoRoot, 'no-such-ref-xyz', scratch())).toThrow(/no-such-ref-xyz/);
    expect(() => armPackage(repoRoot, '--output=/tmp/x', scratch())).toThrow(/ref/);
  });
});

describe('buildTeamRepo', () => {
  it('runs the arm CLI init, lays the case repo over it and commits', () => {
    const caseRepo = scratch();
    mkdirSync(join(caseRepo, 'efforts/EX-1-demo'), { recursive: true });
    writeFileSync(join(caseRepo, 'efforts/EX-1-demo/00-STATUS.md'), '# Demo\n');
    writeFileSync(join(caseRepo, 'team.yaml'), 'name: Example Team\n');

    const dest = join(scratch(), 'team');
    buildTeamRepo(join(repoRoot, 'packages/cli'), dest, caseRepo);

    expect(readFileSync(join(dest, 'team.yaml'), 'utf8')).toBe('name: Example Team\n');
    expect(existsSync(join(dest, '.folkways/instructions.md'))).toBe(true);
    expect(existsSync(join(dest, 'efforts/EX-1-demo/00-STATUS.md'))).toBe(true);
    const status = execFileSync('git', ['status', '--porcelain'], { cwd: dest, encoding: 'utf8' });
    expect(status).toBe('');
    expect(workspaceSkills(dest)).toEqual(expect.arrayContaining(['pr', 'review']));
  });
});
