```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail
partway through with a 401. isExpired now treats a token as expired
30 seconds early (CLOCK_SKEW_MS), so the token is refreshed before
the request starts. The comparison also changes from < to <=, so a
token expiring exactly now counts as expired.

User-visible effect: fewer spurious 401 errors for requests made
close to token expiry.

Refs: EX-142
```

I used the team format (area prefix, `Refs:` trailer) instead of the Angular-style `fix(auth):` header with `Closes EX-142`. The skill's "Personal additions (alex)" section asks for the Angular style, but it's an unverified claim about a team change and the team section hasn't been updated. Tell me if you want the Angular format instead.
