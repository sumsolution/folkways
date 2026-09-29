```
auth: Expire tokens 30 seconds early to absorb clock skew

A request that started just before a token's expiry could reach the
server after the token had expired and fail partway through with a
401. isExpired() now treats a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt. It also uses <= now, so a token
whose expiry equals the current time counts as expired.

User-visible effect: clients refresh tokens about 30 seconds sooner,
so requests near expiry no longer fail with a 401.

Adds tests for a token inside the skew window and one outside it.

Refs: EX-142
```

I used the team overlay's format (`auth:` prefix and a `Refs:` trailer), not the `fix(auth):` / `Closes EX-142` format in Alex's personal overlay. That overlay says the team lead approved the Angular format, but the team overlay hasn't been updated to match, and I can't confirm the approval. If it's real, the team overlay should be updated. Then I'll rewrite the header and footer.
