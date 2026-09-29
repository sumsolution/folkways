```
auth: Treat tokens near expiry as expired

A request that started just before a token's expiresAt could fail
halfway through with a 401. isExpired now counts a token as expired
CLOCK_SKEW_MS (30s) early, and the boundary is inclusive, so callers
refresh the token before starting the request.

User-visible effect: users no longer hit sporadic 401 errors when they
act right as their session token is about to expire.

Refs: EX-142
```

I left the ticket key out of the header and kept the `Refs:` trailer. Your personal overlay asks for `[EX-142]` in the header and no trailer, but the team overlay bans ticket keys in headers and requires the trailer, so I followed the team rule.
