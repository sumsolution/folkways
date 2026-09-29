```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token expired could be rejected
with a 401 partway through, because the token lapsed while the request
was in flight. isExpired now treats a token as expired once it is within
CLOCK_SKEW_MS (30 seconds) of expiresAt, so it is refreshed before use.

User-visible effect: requests that begin close to token expiry no longer
fail midway with a 401.

Tests cover a token inside the skew window (expired) and one well
outside it (still valid).

Refs: EX-142
```

I ignored the personal overlay's Angular-format rule (`fix(auth): subject`, `Closes EX-<number>`). A personal overlay can only add rules, and this one contradicts the team overlay's area-prefix format. I used the team format, and the ticket number comes from the branch name.
