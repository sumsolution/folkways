```
fix(auth): treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS) and uses <= so a token at the boundary is also
expired.

User-visible effect: tokens are refreshed slightly sooner, so users no
longer hit intermittent 401s on requests that straddle expiry.

Closes EX-142

```

I used your personal overlay's Angular header and `Closes EX-142` footer instead of the team's area-prefix format and `Refs:` trailer. I left out the test changes because the body already covers the behavior they check.
