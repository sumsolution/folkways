```
fix(auth): treat tokens within clock skew of expiry as expired

isExpired compared expiresAt to now directly, so a token that was
valid when a request started could expire mid-flight and fail with a
401. Add CLOCK_SKEW_MS (30s) and treat a token as expired once it is
within that window of expiresAt, and at the boundary itself.

User-visible effect: requests made just before a token expires no
longer fail partway through with a 401. The token is refreshed up to
30 seconds earlier instead.

Add tests for a token inside and outside the skew window.

Closes EX-142
```

I used the Angular header and `Closes` footer from your personal overlay, not the team overlay's `auth: Summary` and `Refs:` format. The two overlays conflict, and the personal one says the team overlay is out of date, which I couldn't check.
