```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token's expiry could fail partway
through with a 401. Treat a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt, and use <= so the boundary counts as
expired.

User-visible effect: tokens are refreshed up to 30s sooner, so users
no longer hit intermittent 401s on requests that straddle expiry.

Refs: EX-142
```

I used the team's `<area>: <Summary>` format, not Angular types, and there was no personal overlay to apply. I put the session attribution trailers after `Refs:`, so `Refs:` is not the very last line.
