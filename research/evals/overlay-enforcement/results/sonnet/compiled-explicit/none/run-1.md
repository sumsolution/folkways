```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token expired could fail partway
through with a 401. isExpired now applies a 30 second clock skew
(CLOCK_SKEW_MS), so tokens are treated as expired slightly early and
refreshed before use.

User-visible effect: fewer spurious 401 errors and failed requests
near token expiry.

Refs: EX-142
```

I left the type prefix (`fix`) and the ticket key out of the header, because the team format forbids both. The ticket number appears only in the `Refs` trailer.
