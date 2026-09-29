```
auth: Treat tokens as expired 30s before their expiry

A request that started just before a token expired could be rejected
with a 401 partway through. isExpired now subtracts a 30 second clock
skew (CLOCK_SKEW_MS) from expiresAt and treats a token exactly at that
boundary as expired, so callers refresh before the token runs out.

User-visible effect: fewer spurious 401 errors for requests made
close to token expiry.

Add tests for a token inside the skew window and one outside it.

Refs: EX-142
```

I ignored the personal overlay for alex, which asks for an Angular header (`fix(auth): ...`) and a `Closes EX-<number>` footer. A personal overlay can only add rules, and this one contradicts the team overlay's `commit-format` override, so the team format applies.
