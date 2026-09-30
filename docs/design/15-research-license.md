# 15. Research: the license

**Decision:** Apache-2.0 ([ADR 0004](../adr/0004-license.md)). Corrections made after the research are marked **(checked)**.

**Not legal advice.** This is a research note for the owner's decision. Have a lawyer or your employer's
open-source office confirm before relying on it. Sources were read on 2026-09-30; where a claim comes from
general knowledge rather than a fetched page, it is marked (unverified).

## Recommendation

**Apache-2.0 for everything (code and text assets), one LICENSE. Confidence: medium-high.**
MIT is a close second and would also be defensible (confidence that it is an acceptable fallback: high).

Why Apache-2.0 over MIT:

- It has an explicit patent grant with a retaliation clause ([Apache-2.0 s3](https://www.apache.org/licenses/LICENSE-2.0.txt)).
  MIT has none. That matters most for a framework a company team adopts, and it is the reason to accept a longer text.
- It says the license does not grant trademark rights (s6, same source). That fits a name still awaiting a
  trademark check. MIT is silent (unverified).
- The attribution burden is small if you never add a `NOTICE` file. s4 only requires passing along a NOTICE if one exists.
- It is the license of the closest comparables: Anthropic's skills repo and OpenAI Codex (below).

Why one license, not code plus a docs license: the "docs" here are skills and instructions that agents execute.
They behave like code, and Creative Commons itself [recommends against CC licenses for software](https://creativecommons.org/faq/#can-i-apply-a-creative-commons-license-to-software)
and notes that CC licenses do not address patents. A second license adds a per-file decision to `sync` and to every contributor.

Decide against: GPL/AGPL (bad fit, see table) and MPL-2.0 (file-level copyleft on files that get copied around).

## Comparison

| | MIT | Apache-2.0 | BSD-3-Clause | MPL-2.0 | GPL-3.0 / AGPL-3.0 |
|---|---|---|---|---|---|
| Patent grant | None (unverified) | Yes, with termination on patent suit ([s3](https://www.apache.org/licenses/LICENSE-2.0.txt)) | None (unverified) | Yes (unverified) | GPL-3.0 yes (unverified) |
| Attribution on copied assets | Keep copyright + license text with "substantial portions" | Ship a license copy and any NOTICE; mark modified files ([s4](https://www.apache.org/licenses/LICENSE-2.0.txt)). No NOTICE means little else | As MIT, plus no endorsement clause; must reproduce notice in binaries ([BSD-3 text](https://raw.githubusercontent.com/mvdan/sh/master/LICENSE)) | Only when distributing outside the org; internal use has no obligations ([MPL FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/)) | Copyleft would reach team repos that embed the copied skills if they ever distribute them (unverified) |
| Company allow-lists | Universally allowed (unverified) | Universally allowed (unverified) | Universally allowed (unverified) | Usually allowed with review (unverified) | AGPL is banned outright at some large companies, e.g. [Google](https://opensource.google/documentation/reference/using/agpl-policy) |
| Trademark | Silent | Explicitly not granted (s6) | Bars using names to endorse | Explicitly not granted (unverified) | GPL-3.0 allows the licensor to decline trademark rights (unverified) |
| Contributor friction | Inbound=outbound; DCO or nothing | Same; s5 already says contributions come in under the license | Same | Same | Same |

Text of the licenses: [opensource.org/licenses](https://opensource.org/licenses/) and [choosealicense.com](https://choosealicense.com/licenses/).

## 1. Contributors: DCO, not CLA

Apache-2.0 s5 already licenses contributions to the project on the same terms, so a CLA adds little for a
solo-owned project. The [DCO](https://developercertificate.org/) is a per-commit `Signed-off-by` attesting the
contributor has the right to submit the work. It is cheap, and it is worth asking for because the beta team
works at an employer (see section 4). A CLA is only needed if you might relicense later without asking every
contributor. Spec-kit's CONTRIBUTING [asks for neither](https://raw.githubusercontent.com/github/spec-kit/main/CONTRIBUTING.md).
The repo has no contributors besides you today, so this can wait, but decide before accepting the first outside PR.

## 2. Should the text assets use a different license?

No. What comparable projects do:

| Project | License | Note |
|---|---|---|
| [anthropics/skills](https://github.com/anthropics/skills) | Apache-2.0 for most skills; the docx/pdf/pptx/xlsx skills are source-available, not open source | Skills carry the same license as code. Per-skill `LICENSE.txt` exists, e.g. [skill-creator](https://raw.githubusercontent.com/anthropics/skills/main/skills/skill-creator/LICENSE.txt) (Apache-2.0) |
| [github/spec-kit](https://github.com/github/spec-kit) | MIT | Templates and prompts under the same license as the tool |
| [obra/superpowers](https://raw.githubusercontent.com/obra/superpowers/main/LICENSE) | MIT | A skills collection, one license |
| [openai/codex](https://raw.githubusercontent.com/openai/codex/main/LICENSE) | Apache-2.0 | Agent CLI |

Pattern: agent frameworks and skill collections use one permissive license, MIT or Apache-2.0. None of the
four splits docs from code (from the pages read; I did not audit every subdirectory).
CC-BY-4.0 would add an attribution-with-link duty that suits prose, not files copied into a repo. CC0 is
acceptable for software per the [CC FAQ](https://creativecommons.org/faq/#can-i-apply-a-creative-commons-license-to-software)
but gives up attribution and the trademark carve-out, and some corporate policies treat CC0 cautiously (unverified).

Synced files: no per-file header is needed. Apache-2.0 s4 is met by shipping the LICENSE with the package.
When `sync` copies a skill into a team repo, the team is using the text internally; a redistribution
obligation arises only if they publish it. Optional, cheap: have `sync` write the license identifier in
`.folkways/` manifest metadata so the provenance is machine-readable. No header is required, so I'd skip it.

## 3. Dependencies

| Dependency | License | Effect |
|---|---|---|
| `mvdan.cc/sh/v3` | [BSD-3-Clause](https://raw.githubusercontent.com/mvdan/sh/master/LICENSE) | Compatible with Apache-2.0. Binary distribution must reproduce the copyright notice |
| `gopkg.in/yaml.v3` | [MIT for the libyaml-ported files, Apache-2.0 for the rest](https://raw.githubusercontent.com/go-yaml/yaml/v3/LICENSE) | Compatible. It ships a `NOTICE` (Copyright 2011-2016 Canonical Ltd., Apache-2.0), which the binary must pass on **(checked in the module cache)** |
| Go standard library | BSD-3-Clause | Statically linked into the binary; same notice duty (unverified) |
| npm runtime deps of `@folkways/cli` | None. `packages/cli/package.json` lists only devDependencies | Nothing to declare for npm |

Yes, the Go binary should ship third-party notices. All three licenses above require the copyright notice to
travel with binary redistribution, and `@folkways/hooks-<os>-<cpu>` is a binary redistribution. Add a generated
`THIRD_PARTY_NOTICES` file to each platform package (for example with `go-licenses report`, or a small script
over `go list -m all`; tool choice unverified). The CLI package does not need one while it has no runtime deps.

## 4. Employer IP: check before you publish

Not legal advice. The beta team is at your employer, and you built this while working there or near it
(that is for you and them to establish, not this note). Things to check or ask:

1. Your employment agreement's invention-assignment clause: does it cover work related to the employer's business,
   done on company time or equipment, or using company information?
2. The employer's open-source policy: is a contribution or release of a personal project allowed, is approval
   required, and is there a pre-approval or "no conflict" form?
3. Whether the framework's research used company material. The repo rule says no company content lives here;
   confirm the history is clean too.
4. Whether the beta use by the employer's team needs a written statement that the employer has no claim.
   A short email from the OSS office or legal, kept on file, is what to ask for.
5. Whether the employer wants a specific license, or bars some (see the allow-list row above), and whether any
   employee contributions from the beta team need employer sign-off.
6. State law varies on employee inventions (unverified); a lawyer can say what applies to you.

Do not publish under a license until items 1 and 2 are answered. The `0.0.1` placeholder is `UNLICENSED`,
which grants nothing, so nothing is lost by waiting.

## 5. Mechanics

Files and fields to change once decided:

- `LICENSE` at the repo root: the verbatim Apache-2.0 text (unmodified; the copyright line goes in the appendix
  boilerplate or a README line, e.g. `Copyright 2026 Summer Jordan`).
- A copy of `LICENSE` inside every published package directory: `packages/cli/` and the generated
  `@folkways/hooks-<os>-<cpu>` packages. npm always includes a file named LICENSE/LICENCE in the tarball
  regardless of `files` ([npm docs](https://docs.npmjs.com/cli/v11/configuring-npm/package-json#license)),
  but only if it sits in the package directory. **(checked)** `pnpm pack` in this workspace adds the root
  `LICENSE` to a package that has none, so `@folkways/cli` needs no copy; `pnpm run smoke` asserts it. The
  generated platform packages (hooks PR 5) are packed outside the workspace, so their generator copies it.
- `"license": "Apache-2.0"` (SPDX id, [npm docs](https://docs.npmjs.com/cli/v11/configuring-npm/package-json#license))
  in `packages/cli/package.json`, `packages/hooks/package.json` and `packages/evals/package.json`
  (evals is private and never published; still set it for consistency, or `UNLICENSED` is fine).
- The generated platform packages: whatever script writes their `package.json` (hooks PR 5) must emit `license`,
  the `LICENSE` file and `THIRD_PARTY_NOTICES`. STATUS already lists this as a blocker.
- Go: Go modules have no license field. The binary ships through npm, not as a Go module, so the root
  `LICENSE` is enough. Optionally add `// SPDX-License-Identifier: Apache-2.0` headers
  (not required; skip unless you want them).
- README: a license section (a line, plus a badge if you like). Update `docs/design/00-STATUS.md` (the open
  item at line 82 and the PR 5 note), add an ADR ("0004. License") recording the choice and the employer check,
  and add a changeset for each package whose `license` field changes.
- The unscoped `folkways@0.0.1` placeholder: publish a follow-up version with the new license. Old versions
  stay `UNLICENSED`; that is normal, but npm does not let you edit them.
- Trademark: with Apache-2.0 the name is not licensed. Keep a short line in the README ("The name folkways is not
  covered by the license") once the trademark check is done.

## Open questions for the owner

1. Are you comfortable with the extra length of Apache-2.0 for the patent grant, or do you prefer MIT's brevity?
2. Employer answers (section 4) before any public release.
3. DCO or nothing, decided before the first outside PR.
