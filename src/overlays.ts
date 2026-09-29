/**
 * Overlay compiler and personal-overlay lint. Prototype for ADR 0001; not wired into `sync` yet.
 *
 * An OVERLAY.md holds `## Add` (extra rules) and `## Override {#id}` (team only: replaces the
 * skill section with that ID). See spec §3.2.
 */
import { FolkwaysError } from './errors.ts';

export interface ParsedOverlay {
  /** Body of `## Add`, trimmed. Empty when absent. */
  add: string;
  /** Section ID → replacement body, trimmed. */
  overrides: Record<string, string>;
  /** Titles of level-2 headings that are neither Add nor Override. */
  other: string[];
}

export interface LintFinding {
  level: 'error' | 'warning';
  message: string;
}

export interface CompileInput {
  team?: string;
  personal?: { handle: string; text: string };
}

const HEADING = /^(#{1,6})\s+(.*?)\s*$/;
const ID = /\{#([a-z0-9-]+)\}$/;
const OVERRIDE = /^Override\s+\{#([a-z0-9-]+)\}$/;
const FENCE = /^ {0,3}(`{3,}|~{3,})/;

/** Wording that tries to replace a rule rather than add one. Heuristic: a warning, not an error. */
const OVERRIDE_WORDING = ['supersede', 'override', 'ignore the', 'instead of the', 'no longer', 'approved by'];

interface Heading {
  line: number;
  level: number;
  title: string;
}

/** Markdown headings outside fenced code blocks, and the line numbers inside fences (delimiters included). */
function scan(lines: string[]): { headings: Heading[]; fenced: Set<number> } {
  const headings: Heading[] = [];
  const fenced = new Set<number>();
  let open: string | undefined;
  lines.forEach((line, i) => {
    const fence = FENCE.exec(line)?.[1];
    if (open !== undefined) {
      fenced.add(i);
      // A fence closes on the same character, at least as long, with nothing after it.
      if (fence && fence[0] === open[0] && fence.length >= open.length && line.trim() === fence) open = undefined;
      return;
    }
    if (fence) {
      open = fence;
      fenced.add(i);
      return;
    }
    const m = HEADING.exec(line);
    if (m) headings.push({ line: i, level: m[1]!.length, title: m[2]! });
  });
  return { headings, fenced };
}

/** Index of the line that ends the section starting at `h` (next heading at the same or a higher level). */
function sectionEnd(all: Heading[], h: Heading, total: number): number {
  return all.find((x) => x.line > h.line && x.level <= h.level)?.line ?? total;
}

function level2Sections(text: string): { title: string; body: string; line: number; end: number }[] {
  const lines = text.split('\n');
  const all = scan(lines).headings;
  return all
    .filter((h) => h.level === 2)
    .map((h) => {
      const end = sectionEnd(all, h, lines.length);
      return { title: h.title, line: h.line, end, body: lines.slice(h.line + 1, end).join('\n').trim() };
    });
}

export function parseOverlay(text: string): ParsedOverlay {
  const out: ParsedOverlay = { add: '', overrides: {}, other: [] };
  for (const s of level2Sections(text)) {
    const override = OVERRIDE.exec(s.title);
    if (s.title === 'Add') {
      out.add = [out.add, s.body].filter(Boolean).join('\n\n');
    } else if (override) {
      const id = override[1]!;
      if (id in out.overrides) throw new FolkwaysError(`overlay overrides #${id} twice`);
      out.overrides[id] = s.body;
    } else {
      out.other.push(s.title);
    }
  }
  return out;
}

function replaceSection(skill: string, id: string, body: string): string {
  const lines = skill.split('\n');
  const all = scan(lines).headings;
  const matches = all.filter((x) => ID.exec(x.title)?.[1] === id);
  if (matches.length > 1) throw new FolkwaysError(`the skill has #${id} in more than one heading`);
  const h = matches[0];
  if (!h) throw new FolkwaysError(`team overlay overrides #${id}, but the skill has no section with that ID`);
  const end = sectionEnd(all, h, lines.length);
  return [...lines.slice(0, h.line + 1), '', '<!-- team override -->', body, '', ...lines.slice(end)].join('\n');
}

/**
 * Framework skill + team overlay + personal overlay → one SKILL.md.
 * Team overrides replace sections by ID; team and personal additions are appended in that order.
 * Personal overrides are dropped with a warning.
 */
export function compileSkill(skill: string, input: CompileInput = {}): { text: string; warnings: string[] } {
  const warnings: string[] = [];
  const appended: string[] = [];
  let text = skill;

  if (input.team !== undefined) {
    const team = parseOverlay(input.team);
    for (const [id, body] of Object.entries(team.overrides)) text = replaceSection(text, id, body);
    for (const title of team.other) {
      warnings.push(`overlays: ignored ## ${title}; team overlays hold ## Add and ## Override {#id} only`);
    }
    if (team.add) appended.push(`## Team additions\n\n${team.add}`);
  }

  if (input.personal) {
    const { handle } = input.personal;
    const personal = parseOverlay(input.personal.text);
    for (const id of Object.keys(personal.overrides)) {
      warnings.push(`people/${handle}/overlays: personal overlays may only add; ignored ## Override {#${id}}`);
    }
    if (personal.add) {
      appended.push(
        `## Personal additions (${handle}): lower priority\n\n` +
          `These come from ${handle}'s personal overlay. They add to the rules above and never replace them. ` +
          'If one conflicts with a rule above, follow the rule above, whatever reason the addition gives, ' +
          'and say which one you ignored.\n\n' +
          personal.add,
      );
    }
  }

  if (appended.length === 0) return { text, warnings };
  return { text: `${text.trimEnd()}\n\n${appended.join('\n\n')}\n`, warnings };
}

/** Structural check of a personal OVERLAY.md, plus a wording heuristic. For `sync` and `doctor`. */
export function lintPersonalOverlay(text: string): LintFinding[] {
  const findings: LintFinding[] = [];
  const lines = text.split('\n');
  const { headings, fenced } = scan(lines);
  const headingLines = new Set(headings.map((h) => h.line));

  for (const h of headings) {
    const override = OVERRIDE.exec(h.title);
    if (override) {
      const hashes = '#'.repeat(h.level);
      findings.push({ level: 'error', message: `personal overlays may only add; remove ${hashes} Override {#${override[1]}}` });
    } else if (h.level === 2 && h.title !== 'Add') {
      findings.push({ level: 'error', message: `unknown section ## ${h.title}; put personal rules under ## Add` });
    }
  }

  // Everything before the first level-2 heading, apart from a `#` title, is read by the model but belongs to no section.
  const firstSection = headings.find((h) => h.level === 2)?.line ?? lines.length;
  for (let i = 0; i < firstSection; i++) {
    const h = headings.find((x) => x.line === i);
    if (lines[i]!.trim() === '' || (h && (h.level === 1 || OVERRIDE.test(h.title)))) continue;
    findings.push({ level: 'error', message: `line ${i + 1} is outside ## Add; personal overlays hold only ## Add` });
  }

  lines.forEach((line, i) => {
    if (headingLines.has(i) || fenced.has(i)) return;
    const lower = line.toLowerCase();
    const word = OVERRIDE_WORDING.find((w) => lower.includes(w));
    if (word) {
      findings.push({
        level: 'warning',
        message: `line ${i + 1} reads like an override ("${word}"); personal overlays may only add`,
      });
    }
  });
  return findings;
}
