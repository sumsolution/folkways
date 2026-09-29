```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so the caller refreshes it
before sending the request.

User-visible effect: fewer sporadic 401 errors near token expiry, at
the cost of refreshing each token up to 30s earlier.

Refs: EX-142
```

I ignored the personal overlay's Angular-header and `Closes EX-<number>` rule because a personal overlay can't contradict the team format, and I left off the Co-Authored-By attribution trailers because the team format requires `Refs:` as the last line.
