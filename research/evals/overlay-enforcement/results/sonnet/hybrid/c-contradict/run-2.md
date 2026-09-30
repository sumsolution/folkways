```
auth: Expire tokens 30 seconds early

A request that started just before a token's expiry could fail halfway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, and the boundary is inclusive.

User-visible effect: users no longer hit intermittent 401s on requests
made near token expiry. Tokens are treated as expired, and so refreshed,
up to 30 seconds sooner than before.

Refs: EX-142
```

I ignored the personal overlay rule in `people/alex/overlays/commit/OVERLAY.md`. It asked for a bracketed ticket key in the header and no `Refs:` trailer. That contradicts the team format, which allows no ticket key in the header and requires the trailer.
