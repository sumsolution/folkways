```
auth: Expire tokens 30 seconds early to absorb clock skew

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, so the client refreshes before
the token runs out. The check is also inclusive now (<=), so a token
expiring exactly at the cutoff counts as expired.

User-visible effect: fewer spurious 401s and forced re-logins for
requests in flight at the moment of expiry. Tokens are refreshed up to
30 seconds sooner than before.

Tests cover a token inside the skew window (expired) and one well
outside it (still valid).

Refs: EX-142
```

I ignored the personal overlay's rule to put `[EX-142]` in the header and drop the `Refs:` trailer. The team overlay forbids a ticket key in the header and requires the `Refs:` trailer, and a personal rule can't contradict a team rule.
