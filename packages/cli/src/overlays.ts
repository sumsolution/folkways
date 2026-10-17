/**
 * Team overlay compiler (ADR 0001). `sync` compiles team overlays with `compileSkillFiles`.
 *
 * An OVERLAY.md holds `## Add` (extra rules) and `## Override {#id}` (team only: replaces the
 * skill section with that ID).
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

const HEADING = /^(#{1,6})\s+(.*?)\s*$/;
const ID = /\{#([a-z0-9-]+)\}$/;
const OVERRIDE = /^Override\s+\{#([a-z0-9-]+)\}$/;
const FENCE = /^ {0,3}(`{3,}|~{3,})/;

interface Heading {
  line: number;
  level: number;
  title: string;
}

/** Markdown headings outside fenced code blocks. */
function scan(lines: string[]): { headings: Heading[] } {
  const headings: Heading[] = [];
  let open: string | undefined;
  lines.forEach((line, i) => {
    const fence = FENCE.exec(line)?.[1];
    if (open !== undefined) {
      // A fence closes on the same character, at least as long, with nothing after it.
      if (fence && fence[0] === open[0] && fence.length >= open.length && line.trim() === fence) open = undefined;
      return;
    }
    if (fence) {
      open = fence;
      return;
    }
    const m = HEADING.exec(line);
    if (m) headings.push({ line: i, level: m[1]!.length, title: m[2]! });
  });
  return { headings };
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

/** Headings (outside fences) that carry section `id`. */
function countId(text: string, id: string): number {
  return scan(text.split('\n')).headings.filter((h) => ID.exec(h.title)?.[1] === id).length;
}

/** Throw if one overridden section sits inside another: the outer replacement would remove the inner one. */
function checkNesting(files: Record<string, string>, ids: string[]): void {
  for (const text of Object.values(files)) {
    const lines = text.split('\n');
    const all = scan(lines).headings;
    const sections = all
      .map((h) => ({ h, id: ID.exec(h.title)?.[1] }))
      .filter((s): s is { h: Heading; id: string } => s.id !== undefined && ids.includes(s.id));
    for (const outer of sections) {
      const end = sectionEnd(all, outer.h, lines.length);
      const inner = sections.find((s) => s.h.line > outer.h.line && s.h.line < end);
      if (inner) {
        throw new FolkwaysError(`team overlay overrides both #${outer.id} and #${inner.id}, but #${inner.id} sits inside #${outer.id}; override one`);
      }
    }
  }
}

function unappliedWarnings(overlay: ParsedOverlay): string[] {
  return overlay.other.map((title) => `overlays: ignored ## ${title}; team overlays hold ## Add and ## Override {#id} only`);
}

/**
 * A framework skill folder + a team overlay → the files that change (run by `sync`).
 * `files` maps paths relative to the skill folder to content. An override replaces the section with
 * that ID in whichever markdown file has it (SKILL.md or a reference); team additions go at the end
 * of SKILL.md. An ID no file has, or that more than one heading has, throws.
 */
export function compileSkillFiles(
  files: Record<string, string>,
  team: string,
): { files: Record<string, string>; warnings: string[] } {
  const overlay = parseOverlay(team);
  const out: Record<string, string> = {};
  const current = (path: string): string => out[path] ?? files[path]!;
  const markdown = Object.keys(files).filter((f) => f.endsWith('.md')).sort();

  checkNesting(
    Object.fromEntries(markdown.map((f) => [f, files[f]!])),
    Object.keys(overlay.overrides),
  );
  for (const [id, body] of Object.entries(overlay.overrides)) {
    const holders = markdown.filter((f) => countId(current(f), id) > 0);
    if (holders.length === 0) throw new FolkwaysError(`team overlay overrides #${id}, but the skill has no section with that ID`);
    if (holders.length > 1) throw new FolkwaysError(`the skill has #${id} in more than one file: ${holders.join(', ')}`);
    out[holders[0]!] = replaceSection(current(holders[0]!), id, body);
  }
  if (overlay.add) {
    if (files['SKILL.md'] === undefined) throw new FolkwaysError('team overlay adds rules, but the skill has no SKILL.md');
    out['SKILL.md'] = `${current('SKILL.md').trimEnd()}\n\n## Team additions\n\n${overlay.add}\n`;
  }
  return { files: out, warnings: unappliedWarnings(overlay) };
}
