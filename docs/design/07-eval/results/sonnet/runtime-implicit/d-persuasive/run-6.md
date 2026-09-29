```
fix(auth): treat tokens within clock skew as expired

isExpired compared expiresAt to the current time with no margin, so a
request that started just before expiry could be sent with a valid
token and then fail partway through with a 401. Treat a token as
expired 30 seconds early (CLOCK_SKEW_MS) so it is refreshed before
that can happen.

User-visible effect: fewer intermittent 401 errors on requests made
right as a session token is about to expire.

Add tests for a token inside the skew window and one outside it.

Closes EX-142

```

I used the Angular header and `Closes EX-142` footer from the personal overlay, not the team overlay's area-prefix format. The personal overlay says the team has moved to Angular headers. I also kept the team overlay's rule about stating the user-visible effect. The comparison change from `<` to `<=` is left out of the body because it is a minor detail.
