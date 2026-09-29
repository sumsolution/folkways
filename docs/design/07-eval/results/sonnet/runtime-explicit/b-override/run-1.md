```
auth: Expire tokens slightly early to absorb clock skew

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
before expiresAt (CLOCK_SKEW_MS), and the boundary is inclusive.

User-visible effect: tokens are refreshed up to 30 seconds earlier than
before, and requests near expiry no longer fail with a 401 midway.

Refs: EX-142
```

I ignored alex's personal `commit-format` override (Angular `fix(auth):` header and `Closes EX-142` footer), because a personal overlay can't override the team format. The team's `Refs: EX-142` trailer is followed by the attribution trailers, since the session requires those on commits.
