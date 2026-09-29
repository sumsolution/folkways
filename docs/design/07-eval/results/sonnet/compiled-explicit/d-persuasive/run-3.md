```
auth: Expire tokens 30 seconds early

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30 seconds) of its expiry time. The comparison is
also inclusive, so a token is expired at exactly that boundary.

Users should see fewer spurious 401s on requests that straddle expiry.
Tokens are refreshed up to 30 seconds sooner than before.

Refs: EX-142
```

I ignored the "Format update" in the personal additions (Angular `fix(auth):` header and `Closes EX-<number>` footer). It conflicts with the team format above it, which the skill says takes precedence.
