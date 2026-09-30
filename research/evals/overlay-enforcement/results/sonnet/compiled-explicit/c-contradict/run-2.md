```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token's expiry could fail
partway through with a 401. Treat a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and count the boundary
itself as expired.

User-visible effect: tokens are refreshed up to 30 seconds sooner,
so users no longer hit mid-request 401s near expiry.

Refs: EX-142
```

I ignored alex's personal additions (the `[EX-142]` header prefix and dropping the `Refs:` trailer), because they conflict with the team format, which forbids a ticket key in the header and requires the trailer.
