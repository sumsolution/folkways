#!/usr/bin/env node
import { parseArgs } from 'node:util';
import { FolkwaysError } from './errors.ts';
import { git } from './git.ts';
import { init } from './init.ts';
import { packageVersion } from './package.ts';
import { sync } from './sync.ts';
import { upgrade } from './upgrade.ts';

const USAGE = `usage: folkways <command> [options]

  init <path> [--no-install]   scaffold a team repo at <path>, pin this version, sync, commit
                               (run with: npx @folkways/cli init <path>)
  sync [--force]               copy the installed version's files into this team repo
                               --force  discard edits to framework-owned files
  upgrade <version>            install <version> on branch chore/folkways-<version>, sync, commit
  version                      print the installed version
  hook | review | run | doctor not built yet`;

function teamRoot(): string {
  try {
    return git(process.cwd(), 'rev-parse', '--show-toplevel');
  } catch {
    throw new FolkwaysError('run this inside the team repo');
  }
}

function main(argv: string[]): void {
  const [command, ...rest] = argv;
  switch (command) {
    case 'init': {
      const { positionals, values } = parseArgs({
        args: rest,
        allowPositionals: true,
        options: { 'no-install': { type: 'boolean' } },
      });
      if (positionals.length !== 1) throw new FolkwaysError('usage: folkways init <path> [--no-install]');
      const dest = init(positionals[0]!, { install: !values['no-install'] });
      console.log(`Created ${dest} with folkways ${packageVersion()}.`);
      console.log(`Next: open Claude Code or Copilot CLI in ${dest} and run the setup skill.`);
      return;
    }
    case 'sync': {
      const { values } = parseArgs({ args: rest, options: { force: { type: 'boolean' } } });
      const result = sync(teamRoot(), undefined, { force: values.force });
      if (result.discarded.length > 0) console.log(`Discarded edits:\n${result.discarded.map((e) => `  ${e}`).join('\n')}`);
      for (const w of result.warnings) console.log(`warning: ${w}`);
      const from = result.previous && result.previous !== result.version ? `${result.previous} -> ` : '';
      console.log(`Synced folkways ${from}${result.version}. Review with git status, then commit.`);
      return;
    }
    case 'upgrade': {
      const { positionals } = parseArgs({ args: rest, allowPositionals: true });
      if (positionals.length !== 1) throw new FolkwaysError('usage: folkways upgrade <version>');
      const result = upgrade(process.cwd(), positionals[0]!);
      if (!result.branch) console.log(`Already at ${result.to}; nothing to change.`);
      else console.log(`Committed ${result.from} -> ${result.to} on ${result.branch}. Review, push, open a PR.`);
      return;
    }
    case 'version':
    case '--version':
      console.log(packageVersion());
      return;
    case 'hook':
    case 'review':
    case 'run':
    case 'doctor':
      throw new FolkwaysError(`'${command}' is not built yet (spec §7.2)`);
    case 'help':
    case '--help':
    case '-h':
      console.log(USAGE);
      return;
    default:
      console.error(USAGE);
      process.exit(1);
  }
}

try {
  main(process.argv.slice(2));
} catch (e) {
  if (e instanceof FolkwaysError || (e as { code?: string }).code?.startsWith('ERR_PARSE_ARGS')) {
    console.error(`folkways: ${(e as Error).message}`);
    process.exit(1);
  }
  throw e;
}
