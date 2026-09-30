```
auth: Expire tokens 30 seconds early

A token that was valid when a request started could expire while the
request was in flight, and the request would then fail with a 401.
isExpired now treats a token as expired 30 seconds before expiresAt
(CLOCK_SKEW_MS), so the client refreshes it before sending.

User-visible effect: requests near a token's expiry no longer fail
partway through with a 401. The token is refreshed up to 30 seconds
earlier than before.

Refs: EX-142
```

I left out the Claude co-author and session trailers because the team format requires the `Refs` trailer to be the last line. There was no personal overlay for `alex`, so no personal rules applied.
