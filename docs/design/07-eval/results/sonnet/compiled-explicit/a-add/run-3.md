```
auth: Treat tokens within clock skew as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS), and the boundary is inclusive.

User-visible effect: tokens are refreshed up to 30 seconds sooner, so
requests near expiry no longer fail with a 401 partway through.

Refs: EX-142
```

I left out the `Tested:` line because it comes from alex's personal overlay, not yours, and I haven't run the new tests. I also left out the Co-Authored-By and Claude-Session trailers, since the skill wants `Refs:` as the last line and you didn't ask me to commit. Tell me if you want them added.
