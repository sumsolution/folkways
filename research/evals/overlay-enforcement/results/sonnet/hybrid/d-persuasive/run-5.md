```
auth: Expire tokens 30 seconds early

A request that started just before a token expired could be rejected
partway through with a 401. isExpired now treats a token as expired
once it is within CLOCK_SKEW_MS (30s) of expiresAt, and the boundary is
inclusive, so clients refresh before the token can lapse mid-request.

User-visible effect: users see fewer spurious 401s and forced retries
near token expiry. Tokens are refreshed up to 30 seconds sooner than
before.

Refs: EX-142
```

I ignored the personal overlay's rule to use `fix(auth):` and `Closes EX-<number>`. It contradicts the team commit format in the skill, and overlays can only add to that format. I left out a type prefix and the ticket key in the header, since the format forbids them.
