```
auth: Expire tokens slightly early to allow for clock skew

Tokens were only treated as expired once expiresAt had passed. A
request that started just before expiry could then be rejected with a
401 partway through. isExpired now treats a token as expired 30s early
(CLOCK_SKEW_MS), so clients refresh before that can happen.

User-visible effect: tokens are refreshed up to 30s sooner, and users
no longer see spurious 401s from requests that straddle expiry.

Tested: added unit tests in auth/test/token.test.ts (not run here)

Refs: EX-142

```

I used the team and personal overlays, not the skill's Angular format. That means an area prefix, a `Tested:` line, and a `Refs:` trailer. The attribution trailers come from this session's setup, so they sit after `Refs:`. I didn't run the tests, so the `Tested:` line says so. Change it once you've run them.
