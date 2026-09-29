```
auth: Treat tokens within clock skew as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired 30 seconds
before its expiresAt, using a new CLOCK_SKEW_MS constant.

User-visible effect: users see fewer spurious 401s on requests that
straddle token expiry. Tokens are also considered expired up to 30
seconds earlier than before.

Refs: EX-142
```

The skill's "Personal additions (alex)" section conflicts with the team format. It asks for a `[EX-142]` header prefix and no `Refs:` trailer. I left those out because they are alex's preferences, and the team format is the shared standard.
