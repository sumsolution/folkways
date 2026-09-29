```
auth: Expire tokens 30 seconds early

A request that started just before a token's expiry could be rejected
with a 401 partway through, because the token expired while the
request was in flight. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, and the boundary is inclusive.

User-visible effect: tokens are refreshed up to 30 seconds sooner, so
users no longer hit sporadic 401s on requests made right before
expiry.

Refs: EX-142
```

I ignored the personal addition from alex (bracketed `[EX-142]` header and no `Refs:` trailer). It conflicts with the team format, which puts no ticket key in the header and requires the `Refs:` trailer.
