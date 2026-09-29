```
auth: Expire tokens 30s early to avoid mid-request 401s

A request that started just before a token's expiry could be rejected
with a 401 partway through, because the token expired while the request
was in flight. isExpired now treats a token as expired once it is within
CLOCK_SKEW_MS (30 seconds) of its expiry time, and the comparison is
inclusive.

User-visible effect: tokens are refreshed up to 30 seconds earlier than
before, so users no longer see intermittent 401 failures on requests
that straddle expiry.

Refs: EX-142
```

I put the attribution trailers after `Refs:` because the session's attribution rules require them at the end. That means `Refs: EX-142` is not the very last line, as the commit skill says it should be. Drop the last two lines if you want strict skill format.
