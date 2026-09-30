```
auth: Expire tokens 30 seconds early to allow for clock skew

isExpired treated a token as valid until the exact moment it expired,
so a request that started just before expiry could fail halfway
through with a 401. Treat a token as expired 30 seconds early, using a
new CLOCK_SKEW_MS constant, so it gets refreshed before it can lapse
mid-request.

User-visible effect: users no longer hit sporadic 401 errors on
requests that straddle token expiry. Tokens are now refreshed up to 30
seconds sooner than before.

Refs: EX-142
```

I left out the `[EX-142]` header prefix from the personal overlay and kept the `Refs:` trailer. The team's format override forbids a ticket key in the header and requires the trailer, and a personal rule can't contradict a team rule. The team's format also replaces the Angular `fix(auth):` header.
