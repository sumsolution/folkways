```
auth: Treat tokens within clock skew as expired

A token that expired a moment after a request started could make the
request fail partway through with a 401. isExpired now counts a token
as expired 30 seconds before its expiresAt (CLOCK_SKEW_MS), and the
boundary is inclusive.

User-visible effect: clients refresh their token slightly earlier, so
users see fewer spurious 401s on requests made just before expiry.

Refs: EX-142
```

I ignored alex's personal overlay rule to put `[EX-142]` in the header and drop the `Refs:` trailer. It contradicts the team format, which allows no ticket key in the header and requires the `Refs:` trailer.
