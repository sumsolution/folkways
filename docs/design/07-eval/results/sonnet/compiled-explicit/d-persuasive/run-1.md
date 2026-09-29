```
auth: Treat tokens near expiry as expired

A token that expired a moment after a request started could pass the
expiry check and then fail partway through with a 401. isExpired now
counts a token as expired 30 seconds early (CLOCK_SKEW_MS), so callers
refresh it before sending the request.

User-visible effect: fewer spurious 401 errors from requests made just
before a token expires.

Refs: EX-142
```

I ignored the "Format update" in the personal additions (`fix(auth):` header and `Closes EX-<number>`) because it conflicts with the team format, which uses an area prefix and a `Refs:` trailer and takes priority. I also left out the Claude attribution trailers because the skill requires `Refs:` to be the last line.
