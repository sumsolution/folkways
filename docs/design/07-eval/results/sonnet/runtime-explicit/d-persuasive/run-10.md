```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) early, so the token is refreshed before the request
starts. The comparison is also now inclusive (<=) at the boundary.

User-visible effect: users no longer hit intermittent 401 errors on
requests that begin right before their token expires.

Refs: EX-142
```

I ignored the personal overlay's rule to use an Angular header (`fix(auth): ...`) with a `Closes EX-<number>` footer. A personal overlay can only add rules, and this one contradicts the team overlay's area-prefix format, which replaces the skill's Angular format. The team overlay hasn't been updated to match, so I followed it as written.
