import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { describe, expect, it } from 'vitest';

const REPO = resolve(import.meta.dirname, '..');
const read = (path: string): string => readFileSync(path, 'utf8');

/** Files under `dir`, as sorted POSIX paths relative to it. Empty if `dir` is missing. */
function listFiles(dir: string): string[] {
  if (!existsSync(dir)) return [];
  return readdirSync(dir, { recursive: true, encoding: 'utf8' })
    .filter((rel) => statSync(join(dir, rel)).isFile())
    .map((rel) => rel.split(sep).join('/'))
    .sort();
}

// Rules from docs/skill-authoring.md. Limits from docs/design/09-research-skill-authoring.md.
const SKILLS_DIR = join(REPO, 'skills');
const REQUIRED = ['commit', 'handoff', 'pr', 'review', 'start'];
const NAME = /^[a-z0-9]+(-[a-z0-9]+)*$/;
const NAME_MAX = 64;
const DESCRIPTION_MAX = 500; // spec hard limit is 1,024; framework skills share a listing budget
const SKILL_MD_MAX_LINES = 500;

// Skills that act on the team layout itself: creating `people/<handle>/`, promoting an effort.
const TEAM_ONLY = ['onboard', 'setup', 'status'];

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

/** Body of the `## Config` section, up to the next `## ` heading. */
function configSection(text: string): string | undefined {
  return /^## Config\n([\s\S]*?)(?=^## |(?![\s\S]))/m.exec(text)?.[1];
}

/** The line every `## Config` ends with: the session-start hook supplies the values (spec §3.2). */
const configLine =
  "The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.";

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

    it('has a Config section with the shared line, and no step 0 or personal overlays', () => {
      const config = configSection(text);
      expect(config, 'no "## Config" section').toBeDefined();
      expect(config).toContain(configLine);
      for (const f of markdownFiles(skill)) {
        const body = read(join(SKILLS_DIR, skill, f));
        expect(body, `${f}: step 0`).not.toContain('## Step 0');
        expect(body, `${f}: .local/me/`).not.toContain('.local/me/');
        expect(body, `${f}: personal overlay`).not.toMatch(/personal overlay|<personal root>\/overlays|people\/<handle>\/overlays|OVERLAY\.md\b.*personal/i);
      }
    });

    it('states personal paths against the personal root', () => {
      // Only team-mode procedures name `people/<handle>/`; everything else says `<personal root>/`.
      if (!TEAM_ONLY.includes(skill)) expect(text).not.toContain('people/<handle>');
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

describe('guided setup', () => {
  const setup = read(skillFile('setup'));
  it('asks solo or team and writes mode, commit format and hook toggles', () => {
    for (const word of ['mode: solo', 'solo', 'commit_convention', 'commit_pattern', 'hooks:']) expect(setup).toContain(word);
  });
  it.each(['start', 'onboard'])('%s offers setup when folkways.yaml is missing', (skill) => {
    expect(read(skillFile(skill))).toMatch(/`setup` skill/);
  });
});

describe('solo brain repo', () => {
  // A solo repo is the person's brain repo: no code, so committing a review never moves the reviewed HEAD.
  it('pr has no walk-back over work/ and inbox/ commits', () => {
    expect(read(skillFile('pr'))).not.toMatch(/only `work\/` and `inbox\/`/);
  });
  it.each(['pr', 'commit'])('%s runs the review in a fresh subagent when there is no CLI', (skill) => {
    expect(read(skillFile(skill))).toMatch(/In a solo repo there's no CLI: run the `review` skill in a fresh\s+subagent/);
  });
  it('commit sends code changes to worktrees in a solo repo too', () => {
    const commit = read(skillFile('commit'));
    expect(commit).toMatch(/In the team or solo\s+repo, see \[Where commits go\]/);
    expect(commit).not.toMatch(/commit in place/);
    expect(commit).toMatch(/^\| Solo repo \| `state\/`/m);
  });
  it('setup writes and commits in the brain repo folder, not the session folder', () => {
    const setup = read(skillFile('setup'));
    expect(setup).toMatch(/in the brain repo folder from step 1/);
    expect(setup).toMatch(/git -C <brain repo>/);
  });
  it('setup asks a solo brain repo for its repos', () => {
    expect(read(skillFile('setup'))).toMatch(/^\| Repos \(solo and team\)/m);
  });
  it('setup makes a brain repo for solo, never in a code repo, with projects_dir', () => {
    const setup = read(skillFile('setup'));
    expect(setup).toMatch(/brain repo/);
    expect(setup).toMatch(/never (in )?a code repo/);
    expect(setup).toContain('projects_dir');
  });
  it.each(['start', 'onboard', 'setup'])('%s needs projects_dir in a solo repo', (skill) => {
    expect(read(skillFile(skill))).toContain('`projects_dir` is required');
  });
});

describe('instructions.md', () => {
  const text = read(join(REPO, 'instructions.md'));

  it('points an unconfigured repo to the brain repo', () => {
    expect(text).toMatch(/No `folkways\.yaml`[^]*?brain repo/);
  });

  it('defines the personal root and follows git.commit_convention', () => {
    expect(text).toContain('personal root');
    expect(text).toContain('git.commit_convention');
  });
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
