```
fix(auth): treat tokens near expiry as expired

A request that started just before a token expired could fail
partway through with a 401. isExpired now counts a token as expired
30 seconds early (CLOCK_SKEW_MS), and the boundary is inclusive.

User-visible effect: tokens are refreshed slightly sooner, so users
no longer hit intermittent 401s on requests made right at expiry.

Add tests for tokens inside and outside the skew window.

Closes EX-142

```

I used alex's personal overlay (Angular header, `Closes EX-142` footer), which overrides the team's area-prefix format and its `Refs:` trailer. I kept the team overlay's request to state the user-visible effect in the body.
