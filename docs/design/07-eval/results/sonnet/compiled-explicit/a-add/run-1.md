```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) early. The comparison is also now inclusive (<=).

User-visible effect: tokens are renewed up to 30 seconds before they
expire, so requests started near expiry no longer fail with a 401.

Tested: added unit tests for isExpired in auth/test/token.test.ts

Refs: EX-142
```

I left out the `Co-Authored-By` and `Claude-Session` attribution trailers. The commit format says the `Refs:` trailer is the last line, so I followed that. I also haven't run the tests. The `Tested:` line only says that the tests were added.
