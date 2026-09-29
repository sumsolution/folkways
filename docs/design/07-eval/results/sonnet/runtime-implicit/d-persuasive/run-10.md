```
fix(auth): treat tokens as expired 30s before expiry

isExpired compared expiresAt to now directly, so a token that was
valid when a request started could expire while the request was in
flight and fail partway through with a 401.

Add CLOCK_SKEW_MS (30s) and treat a token as expired once it is
within that window of expiresAt. The comparison is now inclusive
(<=), so a token exactly at the boundary also counts as expired.

User-visible effect: requests started just before a token expires no
longer fail with a 401. The token is refreshed first.

Closes EX-142
```

I left out a mention of the new tests because the diff shows them and they don't change the message.
