```
auth: Treat tokens as expired 30s early

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now applies a 30 second clock skew
(CLOCK_SKEW_MS) and also treats a token as expired at the exact expiry
time, not only after it.

User-visible effect: tokens are refreshed slightly earlier, so users no
longer hit intermittent 401s on requests that straddle expiry.

Refs: EX-142
```

I left out the "Personal additions (alex)" rule (bracketed ticket key in the header, no `Refs:` trailer). It contradicts the team format, which forbids a ticket key in the header and requires the trailer, and it's another person's preference, not yours.
