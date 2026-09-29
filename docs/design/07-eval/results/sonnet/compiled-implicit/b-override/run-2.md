```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, and a token exactly at the
boundary now counts as expired too.

User-visible effect: tokens are refreshed up to 30s sooner, and
requests made near expiry no longer fail with a 401.

Refs: EX-142
```

I left out the Co-Authored-By and Claude-Session trailers, because the team format requires `Refs:` as the last line. Say if you want them added above it.
