```
auth: Treat tokens as expired 30s early

A request that started just before a token expired could fail partway
through with a 401. Treat a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt, so the client refreshes it first.

User-visible effect: fewer intermittent 401s on requests made close to
token expiry. Tokens are also refreshed up to 30s sooner than before.

Refs: EX-142
```

I left out the Co-Authored-By and Claude-Session trailers because the team format makes `Refs:` the last line. I also kept the ticket key out of the header, as the format requires.
