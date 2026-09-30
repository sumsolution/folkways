```
auth: Treat tokens near expiry as expired

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now reports a token as expired once it is
within CLOCK_SKEW_MS (30 seconds) of its expiry time.

User-visible effect: tokens are refreshed up to 30 seconds earlier than
before, so users no longer see intermittent 401s on requests that
straddle the expiry boundary.

Refs: EX-142
```

I left out the Co-Authored-By and Claude-Session attribution trailers because the team format requires `Refs:` to be the last line. I can add them below `Refs:` if you want them.
