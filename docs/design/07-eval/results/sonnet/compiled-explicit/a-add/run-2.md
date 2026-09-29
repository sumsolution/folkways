```
auth: Expire tokens 30s early to avoid mid-request 401s

A request that started just before a token's expiry could be rejected
partway through with a 401. isExpired now treats a token as expired
once it is within CLOCK_SKEW_MS (30s) of expiresAt.

User-visible effect: tokens are refreshed up to 30 seconds sooner, and
requests that begin near expiry no longer fail midway.

Tested: added unit tests in auth/test/token.test.ts (not run here)

Refs: EX-142
```

I followed alex's personal `Tested:` line because it doesn't conflict with the team rules. I wrote "not run here" because I only read the diff and didn't run the tests.
