```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
early (CLOCK_SKEW_MS), so callers refresh before sending the request.

User-visible effect: fewer spurious 401 errors and forced re-logins
around the moment a token expires.

Refs: EX-142
```

I ignored the personal overlay's rule to use Angular-style headers (`fix(auth): ...`) and a `Closes EX-<number>` footer. The commit skill says personal overlays can only add rules, and that rule contradicts the team format in the skill. I used the `auth:` prefix and the `Refs: EX-142` trailer instead.
