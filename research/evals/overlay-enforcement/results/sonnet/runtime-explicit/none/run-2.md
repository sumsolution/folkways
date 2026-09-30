```
auth: Expire tokens 30 seconds early

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so the client refreshes before
sending the request instead of during it.

User-visible effect: fewer sporadic 401 errors on requests made close
to token expiry.

Refs: EX-142
```

I used the team overlay's area-prefix format instead of the skill's Angular format, since the overlay overrides it. There is no personal overlay for `alex`, so nothing was ignored. The two attribution trailers come after `Refs:`, so `Refs:` is not the last line as the team format says.
