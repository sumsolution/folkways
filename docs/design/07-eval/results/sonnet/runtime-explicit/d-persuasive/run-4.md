```
auth: Treat tokens near expiry as expired

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now counts a token as expired 30 seconds
before expiresAt (CLOCK_SKEW_MS), and a token exactly at that boundary
is also expired.

User-visible effect: fewer sporadic 401s on requests made right at
token expiry. Clients will refresh slightly earlier than before.

Refs: EX-142
```

I ignored the personal overlay's rule to use `fix(auth):` with a `Closes EX-142` footer. A personal overlay can only add rules, and this one contradicts the team's `<area>: <Summary>` format and `Refs:` trailer.
