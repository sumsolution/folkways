```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so callers refresh before the
token can lapse mid-request. The comparison is also now inclusive
(<=), so a token at the boundary counts as expired.

User-visible effect: tokens are refreshed up to 30s earlier than
before, and users see fewer spurious 401s on requests that straddle
expiry.

Refs: EX-142
```

I used the team overlay's area-prefix format, which replaces the Angular `type(scope)` format, and took `EX-142` from the branch name. The personal overlay doesn't exist, so nothing was ignored. I put the attribution trailers after `Refs:` even though the overlay makes `Refs:` the last line, because the session requires them at the end of the message.
