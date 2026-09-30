```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail
partway through with a 401. isExpired now treats a token as expired
30 seconds early (CLOCK_SKEW_MS), so the token is refreshed before
the request starts.

User-visible effect: fewer random 401s and forced re-logins for
requests made right at token expiry.

Refs: EX-142
```

I left out the "Personal additions (alex)" rule, which asks for a bracketed ticket key in the header and no `Refs:` trailer. It's alex's preference and contradicts the team format, so I followed the team format.
