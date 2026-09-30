```
auth: Expire tokens 30 seconds early to absorb clock skew

A token that was valid when a request started could expire before the
request finished, so the server rejected it with a 401 partway through.
isExpired now treats a token as expired CLOCK_SKEW_MS (30s) before its
expiresAt, so callers refresh ahead of time instead of failing midway.

User-visible effect: fewer sporadic 401 errors on requests that start
just before a token expires. Tokens are refreshed up to 30 seconds
sooner than before.

Refs: EX-142
```

I ignored the personal overlay's "format update" rule, which asks for Angular `fix(auth):` headers and a `Closes` footer. A personal overlay can't contradict the team format, so I used the team's `auth:` area prefix and `Refs:` trailer. If the team lead has approved a format change, it needs to go into the team overlay.
