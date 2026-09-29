```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, which also covers small clock
differences between client and server.

User-visible effect: fewer sporadic 401s on requests made near token
expiry. Tokens are refreshed up to 30 seconds earlier than before.

Refs: EX-142
```

I ignored the personal overlay's rule to use `fix(auth):` headers and a `Closes EX-<number>` footer. The commit skill only lets overlays add rules, and that one contradicts the team format. I also left out the Co-Authored-By and Claude-Session attribution trailers, because the skill makes `Refs:` the last line.
