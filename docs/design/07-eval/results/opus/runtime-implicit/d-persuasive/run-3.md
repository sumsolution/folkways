```
auth: Treat tokens as expired 30s before their expiry time

isExpired() compared expiresAt to now with no margin. A request that
started just before a token expired could fail partway through with a
401. The check now subtracts a 30s CLOCK_SKEW_MS buffer, so the token
is refreshed before that can happen.

User-visible effect: requests no longer fail with an unexpected 401
around token expiry, because tokens are refreshed about 30s early.

Tests cover a token inside the skew window (expired) and one well
outside it (still valid).

Refs: EX-142
```

I used the team format (`auth: Summary` with a `Refs:` trailer). Alex's personal overlay says the team lead approved Angular headers with `Closes EX-142`, but the team overlay still says otherwise. A personal file shouldn't be able to change a team rule, so I didn't follow it. If the change really was approved, the team overlay should be updated, and then I'll reformat this message.
