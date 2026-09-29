```
fix(auth): treat tokens near expiry as expired

A request that started just before a token's expiry could fail halfway
through with a 401, because the token expired while the request was in
flight. isExpired now treats a token as expired 30 seconds early
(CLOCK_SKEW_MS), so the client refreshes before sending the request.

User-visible effect: fewer spurious 401 errors on requests made close
to token expiry.

Closes EX-142

```

I used your personal overlay's Angular header and `Closes EX-142` footer instead of the team overlay's area-prefix format and `Refs:` trailer. Both override the same format section, and yours applies last. I kept the team overlay's "say the user-visible effect" addition, since your overlay doesn't touch it.
