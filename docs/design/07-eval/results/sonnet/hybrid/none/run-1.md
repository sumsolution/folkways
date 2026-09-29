```
auth: Expire tokens 30s early to allow for clock skew

A token that was valid when a request began could expire while the
request was in flight, so the request failed partway with a 401.
isExpired now treats a token as expired once it is within
CLOCK_SKEW_MS (30s) of its expiry time. The comparison is also
inclusive, so a token at exactly the cutoff counts as expired.

User-visible effect: tokens are refreshed up to 30 seconds sooner, and
requests that start just before expiry no longer fail with a 401.

Refs: EX-142
```

I left out a `fix` type and the ticket key in the header because the team format forbids both. The personal overlay `people/alex/overlays/commit/OVERLAY.md` doesn't exist, so no personal rules applied.
