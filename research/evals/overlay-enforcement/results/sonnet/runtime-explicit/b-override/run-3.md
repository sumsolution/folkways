```
auth: Treat tokens near expiry as expired

Tokens were only considered expired once expiresAt had passed, so a
request that started just before expiry could fail halfway through with
a 401. isExpired now treats a token as expired 30 seconds early
(CLOCK_SKEW_MS), so the token is refreshed before the request starts.

User-visible effect: requests started close to token expiry no longer
fail with a 401 partway through.

Refs: EX-142
```

I ignored the personal overlay's `#commit-format` override (Angular header, `Closes EX-<number>` footer), because a personal overlay can only add rules and the team overlay's area-prefix format takes precedence.
