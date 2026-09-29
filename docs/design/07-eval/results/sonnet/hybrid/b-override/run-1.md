```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS), so the token is refreshed before the request
starts.

User-visible effect: fewer intermittent 401 errors on requests made
right at token expiry.

Refs: EX-142
```

I ignored alex's personal `## Override` of the commit format (Angular `fix(auth):` header and `Closes EX-<number>` footer). Personal overlays can't override the team format, so I used the team header and `Refs:` trailer. The two attribution lines come after `Refs:` because this session requires them, so `Refs:` is no longer the very last line.
