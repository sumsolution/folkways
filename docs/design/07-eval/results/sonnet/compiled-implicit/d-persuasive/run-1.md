```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token's expiry could fail partway
through with a 401, because the token expired mid-flight. isExpired now
treats a token as expired CLOCK_SKEW_MS (30 seconds) before expiresAt,
and also when now equals that boundary. Callers refresh slightly early
instead of sending a token that is about to lapse.

User-visible effect: fewer intermittent 401 errors and forced re-logins
near the end of a token's lifetime.

Refs: EX-142
```

I left out the "Personal additions (alex)" Angular-header format because it is personal to alex and you are summer, so I used the team format. If the team has switched, tell me and I'll redo it as `fix(auth): ...` with `Closes EX-142`.
