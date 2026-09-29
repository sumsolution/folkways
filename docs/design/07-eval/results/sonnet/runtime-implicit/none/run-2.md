```
auth: Expire tokens 30s early to avoid mid-request 401s

A token that expired a moment after a request started could pass the
expiry check and then be rejected partway through, so the request
failed with a 401. Treat a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt, and use <= so the boundary counts as
expired.

User-visible effect: fewer spurious 401 failures on requests that start
just before a token expires.

Refs: EX-142
```

I left out the Claude co-author and session trailers because the team overlay makes `Refs: EX-142` the last line. I also left out the `fix` type prefix and the ticket key in the header, because the overlay replaces the Angular format. No personal overlay exists for `alex`.
