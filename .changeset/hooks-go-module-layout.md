---
"@folkways/hooks": patch
---

Move the Go module to `packages/folkways`, with `cmd/folkways-hook` (shipped, unchanged) and `cmd/folkways`, the Go CLI that will replace the TypeScript one (ADR 0011). `folkways hook <name>` runs the same hooks; `folkways` isn't shipped yet.
