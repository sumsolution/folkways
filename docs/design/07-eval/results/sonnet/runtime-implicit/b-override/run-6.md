```
fix(auth): treat tokens near expiry as expired

isExpired compared expiresAt to the current time with no margin, so a
request that started just before expiry could fail partway through
with a 401. Treat a token as expired once it is within CLOCK_SKEW_MS
(30s) of expiresAt, and use <= so the boundary counts as expired.

User-visible effect: users hit fewer spurious 401s on requests made
right as their token is about to expire, because the token is
refreshed slightly early.

Closes EX-142

```

Alex's personal overlay overrides the team's area-prefix format, so I used the Angular header and `Closes EX-142` footer, and I left out the team's `Refs:` trailer. I kept the team overlay's rule to state the user-visible effect in the body.
