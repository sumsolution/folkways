```
auth: Expire tokens 30s early to absorb clock skew

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, and it also counts a token as
expired when the time equals that boundary.

User-visible effect: fewer spurious 401s on requests made near expiry.
Tokens are refreshed up to 30s earlier than before.

Refs: EX-142
```

I ignored the rule in `people/alex/overlays/commit/OVERLAY.md` that asks for Angular-style headers (`fix(auth):`) and a `Closes EX-<number>` footer. Personal overlays can only add to the team format, and that rule contradicts it. I also left the ticket key out of the header, as the team format requires.
