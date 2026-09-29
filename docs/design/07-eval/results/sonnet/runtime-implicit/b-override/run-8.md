```
fix(auth): treat tokens as expired 30s before expiry

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) early, so the caller refreshes before sending the
request instead of after it has already failed.

User-visible effect: fewer spurious 401s and forced re-logins near
token expiry.

Closes EX-142

```

I used your personal overlay's Angular header and `Closes` footer in place of the team's area-prefix format and `Refs:` trailer. I kept the team overlay's user-visible-effect line in the body, and I left out the tests because the body already covers the behavior.
