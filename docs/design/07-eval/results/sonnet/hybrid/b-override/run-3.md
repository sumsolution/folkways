```
auth: Expire tokens 30s early to allow for clock skew

A request that started just before a token's expiry could fail partway
through with a 401, because the token was only treated as expired once
expiresAt had passed. isExpired now applies a 30 second CLOCK_SKEW_MS
margin, so a token close to expiry is refreshed before the request
starts instead of being rejected during it.

User-visible effect: fewer intermittent 401 errors on requests made
just before a token expires. Tokens are now refreshed up to 30 seconds
earlier than before.

Refs: EX-142
```

I ignored the personal overlay's `## Override` (Angular `fix(auth):` header and `Closes EX-<number>` footer). Personal overrides aren't allowed, and the team format takes precedence.
