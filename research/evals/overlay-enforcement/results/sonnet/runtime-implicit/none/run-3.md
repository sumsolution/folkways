```
auth: Expire tokens 30s early to allow for clock skew

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and the boundary is inclusive,
so the token is refreshed before the request is sent.

Users should see fewer spurious 401s and failed requests around token
expiry.

Refs: EX-142
```

I used the team overlay's format (area prefix, `Refs:` trailer) instead of the skill's Angular default. I put the attribution trailers above `Refs:` so `Refs:` stays the last line, as the overlay requires. `people/alex/overlays/commit/OVERLAY.md` doesn't exist, so no personal overlay applied.
