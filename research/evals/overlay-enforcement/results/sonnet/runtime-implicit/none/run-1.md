```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. Treat a token as expired once it is within 30
seconds of expiry (CLOCK_SKEW_MS), so callers refresh before sending
instead of failing mid-request. The check is now inclusive, so a token
is expired at exactly the boundary.

User-visible effect: fewer spurious 401s and forced sign-outs for
requests made close to token expiry.

Refs: EX-142
```

I left out the Claude co-author and session trailers because the team overlay requires `Refs:` to be the last line. Say if you want them added above it.
