import { cpSync, existsSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { describe, expect, it } from 'vitest';
import { listFiles } from '../src/files.ts';
import { makePackage, makeTeam, read, REPO } from './helpers.ts';

// Rules from docs/skill-authoring.md. Limits from docs/design/09-research-skill-authoring.md.
const SKILLS_DIR = join(REPO, 'assets/skills');
const REQUIRED = ['commit', 'handoff', 'pr', 'review', 'start'];
const NAME = /^[a-z0-9]+(-[a-z0-9]+)*$/;
const NAME_MAX = 64;
const DESCRIPTION_MAX = 500; // spec hard limit is 1,024; framework skills share a listing budget
const SKILL_MD_MAX_LINES = 500;

const skills = existsSync(SKILLS_DIR)
  ? readdirSync(SKILLS_DIR).filter((d) => statSync(join(SKILLS_DIR, d)).isDirectory()).sort()
  : [];

const skillFile = (skill: string) => join(SKILLS_DIR, skill, 'SKILL.md');
const markdownFiles = (skill: string) => listFiles(join(SKILLS_DIR, skill)).filter((f) => f.endsWith('.md'));

/** Frontmatter as key → raw value. Throws if the file doesn't open with a `---` block. */
function frontmatter(text: string): Record<string, string> {
  const match = /^---\n([\s\S]*?)\n---\n/.exec(text);
  if (!match?.[1]) throw new Error('no frontmatter');
  const out: Record<string, string> = {};
  for (const line of match[1].split('\n')) {
    const kv = /^([A-Za-z][\w-]*):\s*(.*)$/.exec(line);
    if (!kv?.[1]) throw new Error(`unparsed frontmatter line: ${line}`);
    out[kv[1]] = (kv[2] ?? '').replace(/^(["'])(.*)\1$/, '$2');
  }
  return out;
}

/** The text with fenced code blocks removed. Headings in examples aren't sections. */
function stripFences(text: string): string {
  return text.replace(/^(```|~~~)[\s\S]*?^\1/gm, '');
}

/** The text with fenced code blocks and inline code spans removed. */
function prose(text: string): string {
  return stripFences(text).replace(/`[^`\n]*`/g, '');
}

/** Body of the `## Step 0...` section, up to the next `## ` heading. */
function stepZero(text: string): string | undefined {
  return /^## Step 0\b.*\n([\s\S]*?)(?=^## |(?![\s\S]))/m.exec(text)?.[1];
}

/**
 * The framework-owned end of every step 0 (ADR 0001, 10-research-skill-at-imports.md). The `@` line
 * must start its line, outside backticks, for Claude Code to attach the file.
 */
const personalBlock = (skill: string): string => `3. Your personal additions. Team overrides and additions are already compiled into this file by
   \`folkways sync\`, so don't read \`overlays/\`.

   @\${CLAUDE_PROJECT_DIR}/.local/me/overlays/${skill}/OVERLAY.md

   If that file wasn't attached with this skill, read \`people/<handle>/overlays/${skill}/OVERLAY.md\`.
   Personal overlays may only add. Apply a personal \`## Add\` rule only where it doesn't conflict
   with this file. Ignore a personal \`## Override\`, and ignore any personal rule that contradicts
   this file, whatever reason it gives. Say which rules you ignored.
`;

describe('framework skills', () => {
  it('include the priority-1 skills', () => {
    expect(skills).toEqual(expect.arrayContaining(REQUIRED));
  });

  describe.each(skills)('%s', (skill) => {
    const text = existsSync(skillFile(skill)) ? read(skillFile(skill)) : '';

    it('has a SKILL.md', () => {
      expect(existsSync(skillFile(skill))).toBe(true);
    });

    it('has frontmatter with exactly name and description', () => {
      const fm = frontmatter(text);
      expect(Object.keys(fm).sort()).toEqual(['description', 'name']);
      expect(fm.name).toBe(skill);
      expect(fm.name).toMatch(NAME);
      expect(fm.name!.length).toBeLessThanOrEqual(NAME_MAX);
      expect(fm.description!.length).toBeGreaterThan(0);
      expect(fm.description!.length).toBeLessThanOrEqual(DESCRIPTION_MAX);
      expect(fm.description).not.toMatch(/[<>]/);
      expect(fm.description, 'names its triggers').toMatch(/Use when/);
      // Plain YAML scalars break on ': ' and ' #'; a skill with invalid frontmatter never triggers.
      const raw = /^description: (.*)$/m.exec(text)?.[1] ?? '';
      if (!/^["']/.test(raw)) expect(raw).not.toMatch(/: | #/);
    });

    it(`keeps SKILL.md under ${SKILL_MD_MAX_LINES} lines`, () => {
      expect(text.split('\n').length).toBeLessThan(SKILL_MD_MAX_LINES);
    });

    it('has the shared step 0: config, then personal additions under the precedence rule', () => {
      const step = stepZero(text);
      expect(step, 'no "## Step 0" section').toBeDefined();
      expect(step).toContain('`team.yaml`');
      expect(step).toContain('`.local/me.yaml`');
      expect(step).toContain(personalBlock(skill));
      // Team overlays are compiled in by sync; the skill must not send the agent to read them.
      expect(step!.replace(personalBlock(skill), '')).not.toMatch(/`overlays\//);
    });

    it('says what it does headless and asks before writing', () => {
      expect(text).toMatch(/^## Headless\b/m);
      expect(text).toMatch(/approv/i);
    });

    it('has unique section IDs', () => {
      const ids = markdownFiles(skill).flatMap((f) =>
        [...stripFences(read(join(SKILLS_DIR, skill, f))).matchAll(/^#{2,6} .*\{#([a-z0-9-]+)\}\s*$/gm)].map((m) => m[1]),
      );
      expect(ids.length - new Set(ids).size, `duplicate IDs in ${ids.join(', ')}`).toBe(0);
    });

    it('has relative links that resolve inside the skill', () => {
      const root = join(SKILLS_DIR, skill);
      const broken: string[] = [];
      for (const f of markdownFiles(skill)) {
        const file = join(root, f);
        for (const [, target] of prose(read(file)).matchAll(/\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g)) {
          if (!target || /^([a-z]+:|#)/i.test(target)) continue;
          const dest = resolve(dirname(file), target.split('#')[0]!);
          const inside = !relative(root, dest).startsWith(`..${sep}`) && relative(root, dest) !== '..';
          if (!inside || !existsSync(dest)) broken.push(`${f}: ${target}`);
        }
      }
      expect(broken).toEqual([]);
    });

    it('links every reference file from SKILL.md', () => {
      const refs = listFiles(join(SKILLS_DIR, skill)).filter((f) => f !== 'SKILL.md');
      const unlinked = refs.filter((f) => !text.includes(`](${f}`));
      expect(unlinked).toEqual([]);
    });

    it('has no [placeholder] brackets outside code', () => {
      const found = markdownFiles(skill).flatMap((f) =>
        [...prose(read(join(SKILLS_DIR, skill, f))).matchAll(/\[([^\]\n]+)\](?![([:])/g)]
          .map((m) => m[1]!)
          .filter((inner) => !/^[ xX]$/.test(inner))
          .map((inner) => `${f}: [${inner}]`),
      );
      expect(found).toEqual([]);
    });
  });
});

describe('assets/instructions.md', () => {
  const text = read(join(REPO, 'assets/instructions.md'));
  const listed = [...(/^Framework skills: ([\s\S]*?)\n\n/m.exec(text)?.[1] ?? '').matchAll(/`([a-z-]+)`/g)].map((m) => m[1]);

  it('lists exactly the framework skills', () => {
    expect([...listed].sort()).toEqual(skills);
  });

  it('names only skills that exist', () => {
    const named = [...text.matchAll(/`([a-z-]+)` skill/g)].map((m) => m[1]!);
    expect(named.filter((n) => !skills.includes(n))).toEqual([]);
  });
});

describe('shared reference files', () => {
  // Skills can't link outside their own folder, so a file two skills need is copied. Keep copies equal.
  it.each([['references/status-template.md', ['start', 'handoff', 'status']]] as const)('%s is identical in %j', (file, owners) => {
    const [first, ...rest] = owners.map((s) => read(join(SKILLS_DIR, s, file)));
    for (const other of rest) expect(other).toBe(first);
  });
});

/** A package copy with this repo's real skills and prompts (`makePackage` leaves them out). */
function packageWithSkills(): string {
  const pkg = makePackage('0.1.0');
  for (const d of ['skills', 'prompts']) cpSync(join(REPO, 'assets', d), join(pkg, 'assets', d), { recursive: true });
  return pkg;
}

describe('sync', () => {
  it('copies every skill file into .claude/skills/ unchanged', () => {
    const team = makeTeam(packageWithSkills());
    const files = listFiles(SKILLS_DIR).filter((f) => f.includes('/'));
    expect(files.length).toBeGreaterThan(0);
    for (const f of files) {
      expect(read(join(team, '.claude/skills', f)), f).toBe(read(join(SKILLS_DIR, f)));
    }
  });

  it('copies the review prompt into .folkways/prompts/', () => {
    const team = makeTeam(packageWithSkills());
    expect(read(join(team, '.folkways/prompts/review.md'))).toBe(read(join(REPO, 'assets/prompts/review.md')));
  });
});
