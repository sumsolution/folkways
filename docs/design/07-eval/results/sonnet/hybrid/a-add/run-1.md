```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and the boundary is inclusive.

User-visible effect: tokens are refreshed up to 30 seconds earlier, so
users no longer hit intermittent 401s on requests that straddle expiry.

Tested: added unit tests for isExpired on both sides of the clock skew
window

Refs: EX-142
```

I applied alex's personal `## Add` rule (the `Tested:` line before the trailer) because it doesn't conflict with the skill, and I ignored no rules. I didn't run the tests, so the `Tested:` line only says the unit tests were added. Edit it if you ran them.
