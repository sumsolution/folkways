```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS), so callers refresh before the request starts.

User-visible effect: fewer intermittent 401s near token expiry, at the
cost of refreshing tokens up to 30 seconds sooner.

Tested: added unit tests for isExpired in auth/test/token.test.ts (not run)

Refs: EX-142
```

I applied your overlay's `Tested:` line, and it didn't conflict with the skill. I left out the Claude attribution trailers because the skill makes `Refs:` the last line, and I didn't run the tests, so the `Tested:` line says so.
