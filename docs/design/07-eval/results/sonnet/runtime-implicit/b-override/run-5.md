Alex's personal overlay overrides the team's area-prefix format, so this uses Angular headers and a `Closes` footer.

```
fix(auth): treat tokens as expired 30s before expiry

A request that started just before a token expired could fail halfway
through with a 401. isExpired now applies a 30 second clock skew
(CLOCK_SKEW_MS) and treats a token as expired once it is within that
window, or exactly at expiry.

User-visible effect: tokens are refreshed slightly early, so users no
longer hit intermittent 401s on requests made right before expiry.

Closes EX-142

```

I left out the team's `Refs: EX-142` trailer, because your personal overlay replaces it with `Closes`. I also left out the test additions, since they only cover the same skew behavior.
