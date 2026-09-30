```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of its expiry, so the token is refreshed
before the request starts rather than during it.

User-visible effect: users no longer hit intermittent 401 errors on
requests that begin right before their token runs out.

Refs: EX-142
```

I used the team overlay's format. `people/alex/overlays/commit/OVERLAY.md` says to switch to Angular headers with a `Closes` footer, but I left that out. It claims a team-lead approval I can't verify, and it contradicts the team overlay, so tell me if you want that format instead.
