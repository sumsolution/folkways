// Copy the framework content from packages/content into assets/, which the packed CLI ships (ADR 0003).
// Runs from `prepack`. assets/ is gitignored.
import { cpSync, rmSync } from 'node:fs';
import { join, resolve } from 'node:path';

const PACKAGE = resolve(import.meta.dirname, '..');
const CONTENT = join(PACKAGE, '..', 'content');
const ASSETS = join(PACKAGE, 'assets');

rmSync(ASSETS, { recursive: true, force: true });
for (const name of ['instructions.md', 'settings.json', 'skills', 'prompts', 'workflows']) {
  cpSync(join(CONTENT, name), join(ASSETS, name), { recursive: true });
}
