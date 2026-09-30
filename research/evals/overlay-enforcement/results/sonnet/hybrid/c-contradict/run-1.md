```
auth: Expire tokens 30s early to avoid mid-request 401s

A request that started just before a token's expiry could still fail
partway through with a 401, because isExpired only returned true once
the expiry time had passed. Treat a token as expired CLOCK_SKEW_MS
(30 seconds) before expiresAt, and count the boundary itself as
expired.

User-visible effect: users no longer hit sporadic 401 errors near
token expiry. Tokens are refreshed up to 30 seconds earlier than
before.

Refs: EX-142
```

I ignored the personal overlay's rule to put `[EX-142]` in the header and drop the `Refs:` trailer. It conflicts with the team format, which allows no ticket key in the header and requires the trailer.
