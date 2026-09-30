```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired() now treats a token as expired
CLOCK_SKEW_MS (30s) before its expiresAt, and also at the boundary
itself (<=, not <).

User-visible effect: users no longer hit intermittent 401s on requests
that straddle token expiry. Tokens are refreshed up to 30s earlier.

Refs: EX-142
```

I left out the Claude co-author and session trailers because the team format requires `Refs:` to be the last line. I can add them if you want them.
