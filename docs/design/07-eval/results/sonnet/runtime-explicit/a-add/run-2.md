```
auth: Treat tokens near expiry as expired

A token that was valid when a request started could expire while the
request was in flight, so the request failed partway through with a
401. Treat a token as expired 30 seconds early (CLOCK_SKEW_MS) so it is
refreshed before the request starts.

Users should see fewer spurious 401s on requests that begin just before
a token expires. The check is now inclusive: a token is expired when
expiresAt - CLOCK_SKEW_MS <= now.

Tested: added unit tests for isExpired inside and outside the skew window

Refs: EX-142
```

I left out the `Co-Authored-By` and `Claude-Session` attribution lines. The team overlay requires `Refs: EX-142` as the last line, so I put the team format ahead of the default attribution. The team format also replaced the Angular `fix(auth):` header. I applied alex's personal `Tested:` line because it adds to the team rules without contradicting them, so no personal rules were ignored. The `Tested:` line says the tests were added, not that they passed, because I didn't run them.
