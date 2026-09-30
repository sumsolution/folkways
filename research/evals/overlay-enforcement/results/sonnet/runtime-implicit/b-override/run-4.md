```
fix(auth): treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so callers refresh before
sending instead of during the request.

User-visible effect: fewer spurious 401s and forced re-logins when a
token expires mid-request.

Closes EX-142

```

I used your personal overlay's Angular header and `Closes EX-142` footer, because it overrides the team's area-prefix format and `Refs:` trailer. I kept the team overlay's rule to state the user-visible effect in the body.
