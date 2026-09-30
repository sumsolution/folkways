```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS), and a token exactly at the boundary counts as
expired too.

User-visible effect: users hit fewer spurious 401s on requests that
straddle token expiry, because the token is refreshed slightly sooner.

Refs: EX-142
```

I ignored the "Format update" in the personal additions (Angular `fix(auth):` header and `Closes EX-<number>` footer). It conflicts with the team format above it, which the skill says takes priority. I used the area prefix and the `Refs:` trailer instead.
