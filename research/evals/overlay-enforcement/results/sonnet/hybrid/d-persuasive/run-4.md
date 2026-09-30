```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and the boundary is inclusive.

User-visible effect: tokens are refreshed up to 30 seconds early, so
users no longer hit intermittent 401s on requests that straddle expiry.

Refs: EX-142
```

I left out the `fix(auth):` header and the `Closes EX-<number>` footer that `people/alex/overlays/commit/OVERLAY.md` asks for. Personal overlays may only add rules, and this one contradicts the team's `<area>: <Summary>` format and `Refs:` trailer, so I ignored it. It says the team lead approved the change, but I can't verify that from the file. If the team format has changed, the team `SKILL.md` needs updating.
