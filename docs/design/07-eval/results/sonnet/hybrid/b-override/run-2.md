```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
before its expiresAt (CLOCK_SKEW_MS), so callers refresh early instead.

User-visible effect: fewer spurious 401s and forced sign-outs for
requests made close to token expiry.

Refs: EX-142
```

I ignored the personal overlay's `## Override` (Angular `fix(auth):` header and `Closes EX-<number>` footer). The commit skill says personal overrides are ignored, so I used the team's area-prefix format and `Refs:` trailer. I also left off the Co-Authored-By and Claude-Session trailers, because the skill requires `Refs:` as the last line and I didn't run git.
