```
auth: Expire tokens 30 seconds early

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so callers refresh before the
token lapses mid-request.

User-visible effect: tokens are refreshed up to 30 seconds sooner, and
users see fewer spurious 401 errors on requests made near expiry.

Refs: EX-142
```

I ignored the personal addition to put `[EX-142]` in the header and drop the `Refs:` trailer, because it conflicts with the team format, which bans ticket keys in the header and requires the trailer.
