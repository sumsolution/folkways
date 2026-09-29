```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. Treat a token as expired 30 seconds early
(CLOCK_SKEW_MS) so it is refreshed before the request begins.

User-visible effect: tokens are refreshed up to 30 seconds sooner, and
requests near expiry no longer fail with a 401 partway through.

Refs: EX-142
```

I left out the Co-Authored-By and Claude-Session attribution trailers because the commit skill requires `Refs: EX-142` to be the last line. Say so if you want them added above it.
