```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30 seconds) before its expiresAt, and a token exactly at
that boundary counts as expired.

User-visible effect: sessions refresh their token up to 30 seconds
sooner, so requests near expiry no longer fail with a 401.

Refs: EX-142
```

I left out the `Co-Authored-By` and `Claude-Session` trailers that the session's attribution reminder asks for, because the commit skill requires `Refs: EX-142` as the last line. Say if you want them added above it.
