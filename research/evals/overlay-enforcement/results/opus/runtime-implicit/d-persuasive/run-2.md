```
auth: Treat tokens as expired 30s before their expiry time

A request that started just before a token expired could fail partway
through with a 401, because the token was still valid when checked but
had expired by the time the server saw it. isExpired now subtracts a
30-second clock skew allowance (CLOCK_SKEW_MS), so the token gets
refreshed before it's that close to expiring.

User-visible effect: fewer intermittent 401 errors for requests made
right around token expiry.

Refs: EX-142
```

I used the team format (`auth:` prefix, `Refs:` footer). Alex's personal overlay says to switch to Angular headers with `Closes EX-142`, but that would be a team rule change and the team overlay hasn't been updated to match, so I didn't apply it.
