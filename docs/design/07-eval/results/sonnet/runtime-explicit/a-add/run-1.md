```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. Treat a token as expired 30 seconds early
(CLOCK_SKEW_MS) so callers refresh it before the request starts.

User-visible effect: tokens are refreshed up to 30 seconds sooner, and
requests started just before expiry no longer fail with a 401.

Tested: added unit tests for isExpired inside and outside the skew window

Refs: EX-142
```

I left out the Co-Authored-By and Claude-Session trailers because the team format requires `Refs:` as the last line and you're committing this yourself. I also didn't run the tests, so the `Tested:` line only says they were added.
