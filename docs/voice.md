# Voice

How folkways docs sound. One page, so you can hold it in your head while you write. Summer owns this;
tune it freely.

## In one breath

Direct, succinct, warm, sometimes witty. We write like a senior engineer explaining the system to a
teammate they like: precise first, friendly always, a joke only when it's free.

## Principles

1. **Clarity first.** The technical content stands on its own. Personality is seasoning. If a joke costs
   clarity, cut the joke.
2. **Short and plain.** Short sentences, common words. One idea per paragraph. Tables for anything with
   rows.
3. **Concrete beats clever.** Name the file, the command, the mechanism. "Sync stops if a framework file's
   checksum changed" beats "sync keeps things safe".
4. **"We", not "you must".** folkways is about teams. Review gates, evidence and asking before pushing are
   things we do because they're right, not rules to fear.
5. **Honest about status.** Say what's built and what isn't. Mark unbuilt things "coming soon". Never
   describe a plan in the present tense.
6. **Delight on purpose.** At most two emoji or kaomoji per page, where something genuinely good happens.
   At most one or two cultural references per page, unexplained.

## Do and don't

| Do | Don't |
|---|---|
| "`sync` stops and lists the edited files." | "`sync` intelligently protects your customizations." |
| "We review before we push." | "You MUST NEVER push without review!!" |
| "Coming soon: the `setup` skill." | "Run `setup` to configure everything." (it isn't built) |
| One 🙌 when the team ships something together | 🚀✨🔥 on every heading |
| `<(^_^<)` once, when a team repo is ready | Kaomoji in tables, code blocks or error text |
| A quiet nod: "set and setting" for context | A paragraph explaining the reference |
| Plain words: "copies", "checks", "stops" | Hype: "revolutionary", "supercharge", "seamless", "unlock" |
| One metaphor, if it clarifies | Metaphors stacked on metaphors |

## Our emoji

| Mark | Means | Use for |
|---|---|---|
| 🙌 | High five, working together | A shared win, a teammate's contribution, collaboration |
| `<(^_^<)` `(>^_^)>` `<(^_^)>` | Kirby dancing: pure delight | Something is ready and it's genuinely exciting |

Nothing else, unless it earns its place. Never in reference tables, command blocks or error messages.

## References

Easter eggs, not themes. They fit because they're true, not because they're decoration.

- **Philosophy and sociology.** Sumner, folkways and mores are the name. Use them where they explain the design.
- **Set and setting.** Context shapes the experience. That's context engineering, and it fits once, where
  we talk about what an agent reads at session start.
- **PLUR** (peace, love, unity, respect). A light touch where we talk about working together.
- Never drug jokes. Never explain the reference.

## Before and after

**1. Describing sync**

> Before: folkways features a revolutionary sync engine that seamlessly keeps your team's AI setup in
> perfect harmony, so you never have to worry about anything ever again! 🚀✨

> After: `folkways sync` copies the framework's text files into your team repo and records a checksum
> for each. If someone edited one, sync stops and tells you which, so nothing gets clobbered.

**2. The review gate**

> Before: Agents must never, under any circumstances, push code that has not been reviewed. Violations
> will be blocked.

> After: Before a PR opens, a fresh agent that never saw the work reviews it. The author can't frame
> the review, so it catches what the author's context hides. We all do it; it's how we trust each
> other's code. 🙌

**3. Unbuilt features**

> Before: Run `folkways doctor` to diagnose and fix your installation.

> After: `folkways doctor` (coming soon) will check config, hooks and version drift and print fixes.
> Today it says it isn't built yet.

## Checklist before you commit a page

- [ ] Every command runs, or is marked "coming soon".
- [ ] Nothing described as built that isn't.
- [ ] No company-specific names, URLs or keys. Examples use `Example Team`, `EX`, `example.com`.
- [ ] Two or fewer emoji or kaomoji. One or two references.
- [ ] Readable in under five minutes.
- [ ] Read it aloud once. Cut anything that sounds like a pitch deck.
