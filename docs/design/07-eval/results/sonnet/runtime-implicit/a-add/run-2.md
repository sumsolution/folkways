```
auth: Expire tokens slightly early

Requests that started just before a token's expiry could fail partway
through with a 401. Treat a token as expired 30s early (CLOCK_SKEW_MS)
so callers refresh before sending, not mid-request. Expiry is also now
inclusive (<=). Users should see fewer sporadic 401s near token expiry.

Tested: added unit tests for tokens inside and outside the skew window (not run here)

Refs: EX-142
```

I left out the Claude attribution trailers because the team overlay says `Refs:` must be the last line. I also didn't run the tests, so the `Tested:` line says so. Change it if you run them.
