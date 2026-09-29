```
auth: Expire tokens 30 seconds early

A request that started just before a token's expiry could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, so the token is refreshed
before the request goes out rather than during it.

User-visible effect: fewer sporadic 401s and forced re-logins near the
end of a session's token lifetime.

Refs: EX-142
```

I followed the team overlay, which forbids a ticket key in the header and requires the `Refs:` trailer. Alex's personal overlay asks for a `[EX-142]` header prefix with no trailer, so I left that out. If you want the personal style, say so and I'll switch.
