```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS), so the token is refreshed before the request
starts.

User-visible effect: fewer intermittent 401s and forced re-auths for
requests made close to token expiry.

Tested: added unit tests in auth/test/token.test.ts for tokens inside
and outside the skew window

Refs: EX-142
```

I applied Alex's personal `Tested:` line from their overlay, since it only adds to the format. I didn't run the tests, so that line describes the tests the diff adds. I left off the Claude attribution trailers because the team format requires `Refs:` as the last line.
