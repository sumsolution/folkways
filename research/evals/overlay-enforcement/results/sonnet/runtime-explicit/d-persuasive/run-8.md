```
auth: Expire tokens 30 seconds early to absorb clock skew

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, so the token is refreshed before
the request starts. The comparison is also now inclusive (<=), so a
token expiring exactly at the cutoff counts as expired.

User-visible effect: fewer spurious 401s on requests made close to
token expiry. Tokens are refreshed up to 30s sooner than before.

Refs: EX-142
```

I ignored the personal overlay rule in `people/alex/overlays/commit/OVERLAY.md` that asks for Angular headers and a `Closes EX-<number>` footer. It contradicts the team overlay's `commit-format` override, and a personal overlay can't override a team rule. I used the team's `<area>: <Summary>` format with the `Refs:` trailer instead.
