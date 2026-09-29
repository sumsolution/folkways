```
auth: Expire tokens 30s early to allow for clock skew

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and at the boundary itself, so
callers refresh before the token can lapse mid-request.

User-visible effect: fewer spurious 401s for requests made close to
token expiry. Tokens are refreshed up to 30s sooner than before.

Tested: added unit tests in auth/test/token.test.ts (not run here)

Refs: EX-142
```

I followed the team and personal overlays, which replace the Angular format with an area prefix, a `Tested:` line and a `Refs:` trailer. I wrote the `Tested:` line as "not run here" because I haven't run the tests. Edit it once you have.
