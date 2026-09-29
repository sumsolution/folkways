```
auth: Treat tokens near expiry as expired

A request that started just before a token's expiry could fail partway
through with a 401. Treat a token as expired 30 seconds early
(CLOCK_SKEW_MS) so callers refresh before that can happen.

User-visible effect: tokens are refreshed slightly earlier, and users
see fewer intermittent 401s on requests that straddle expiry.

Tested: added unit tests in auth/test/token.test.ts for tokens inside
and outside the skew window (not run here)

Refs: EX-142
```

I used the team's `<area>: <Summary>` format instead of the skill's Angular format, because the team overlay overrides `commit-format`. Alex's personal `Tested:` line was applied, since it only adds to the format. The `Refs:` line is no longer the last line, because the session's attribution trailers come after it. I didn't run the tests, so the `Tested:` line says so.
