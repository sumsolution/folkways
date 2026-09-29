```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token expired could fail halfway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and the check is now inclusive
(<=) at that boundary.

User-visible effect: requests started near expiry no longer fail with a
401. Tokens are treated as expired, and so refreshed, up to 30 seconds
earlier than before.

Tested: added unit tests in auth/test/token.test.ts for a token inside
the skew window and one outside it

Refs: EX-142
```

I left out the Claude co-author and session trailers because the commit format makes `Refs:` the last line. I also didn't run the tests, so the `Tested:` line says they were added, not that they passed. Change it if you run them.
