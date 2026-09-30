```
fix(auth): treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and the boundary is inclusive.

User-visible effect: users no longer see intermittent 401 errors on
requests that begin right before their token expires. The client
refreshes the token slightly earlier instead.

Tests cover a token inside the skew window (expired) and one well
outside it (still valid).

Closes EX-142
```

I used your personal overlay's Angular header and `Closes EX-142` footer, which override the team's area-prefix format and `Refs:` trailer. I kept the team overlay's rule to state the user-visible effect in the body.
