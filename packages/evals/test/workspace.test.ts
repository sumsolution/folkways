import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import { armCli, buildTeamRepo, WORKTREE, workspaceSkills } from '../src/workspace.ts';

const repoRoot = resolve(import.meta.dirname, '../../..');
const scratch = () => mkdtempSync(join(tmpdir(), 'evals-ws-'));

describe('armCli', () => {
  it('builds the working tree CLI for WORKTREE', () => {
    const cli = armCli(repoRoot, WORKTREE, scratch());
    const version = JSON.parse(readFileSync(join(repoRoot, '.release-please-manifest.json'), 'utf8'))['.'];
    expect(execFileSync(cli, ['version'], { encoding: 'utf8' })).toBe(`${version}\n`);
  });

  it('builds the CLI at a git ref, once', () => {
    const dir = scratch();
    const cli = armCli(repoRoot, 'HEAD', dir);
    expect(execFileSync(cli, ['version'], { encoding: 'utf8' })).toMatch(/^\d+\.\d+\.\d+/);
    expect(armCli(repoRoot, 'HEAD', dir)).toBe(cli);
  });

  it('refuses a ref from before the Go CLI had init', () => {
    const repo = scratch();
    mkdirSync(join(repo, 'packages/cli/src'), { recursive: true });
    writeFileSync(join(repo, 'packages/cli/src/cli.ts'), '');
    const env = { ...process.env, GIT_CONFIG_GLOBAL: '/dev/null', GIT_AUTHOR_NAME: 't', GIT_AUTHOR_EMAIL: 't@example.com', GIT_COMMITTER_NAME: 't', GIT_COMMITTER_EMAIL: 't@example.com' };
    for (const args of [['init', '-q'], ['add', '-A'], ['commit', '-q', '-m', 'old']]) execFileSync('git', args, { cwd: repo, env });
    expect(() => armCli(repo, 'HEAD', scratch())).toThrow(/no Go CLI init; arms need 141a6f4 or later/);
  });

  it('rejects a ref that is not a commit', () => {
    expect(() => armCli(repoRoot, 'no-such-ref-xyz', scratch())).toThrow(/no-such-ref-xyz/);
    expect(() => armCli(repoRoot, '--output=/tmp/x', scratch())).toThrow(/ref/);
  });
});

describe('buildTeamRepo', () => {
  it('runs the arm CLI init, lays the case repo over it and commits', () => {
    const caseRepo = scratch();
    mkdirSync(join(caseRepo, 'efforts/EX-1-demo'), { recursive: true });
    writeFileSync(join(caseRepo, 'efforts/EX-1-demo/00-STATUS.md'), '# Demo\n');
    writeFileSync(join(caseRepo, 'team.yaml'), 'name: Example Team\n');

    const dest = join(scratch(), 'team');
    buildTeamRepo(armCli(repoRoot, WORKTREE, scratch()), dest, caseRepo);

    expect(readFileSync(join(dest, 'team.yaml'), 'utf8')).toBe('name: Example Team\n');
    expect(existsSync(join(dest, '.folkways/instructions.md'))).toBe(true);
    expect(existsSync(join(dest, 'efforts/EX-1-demo/00-STATUS.md'))).toBe(true);
    const status = execFileSync('git', ['status', '--porcelain'], { cwd: dest, encoding: 'utf8' });
    expect(status).toBe('');
    expect(workspaceSkills(dest)).toEqual(expect.arrayContaining(['pr', 'review']));
  });
});
