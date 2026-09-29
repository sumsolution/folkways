```
auth: Expire tokens 30 seconds early

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so callers refresh before
sending instead of racing the expiry. The comparison is also inclusive
now, so a token at exactly the cutoff counts as expired.

User-visible effect: fewer spurious 401s on requests made close to
token expiry. Tokens are refreshed up to 30 seconds sooner than before.

Refs: EX-142
```

I ignored the personal overlay's rule in `people/alex/overlays/commit/OVERLAY.md`, which asked for `fix(auth):` headers and a `Closes EX-<number>` footer. The commit skill says personal overlays can only add and must not contradict the team format, so I used the team's `<area>: <Summary>` header and `Refs:` trailer.
