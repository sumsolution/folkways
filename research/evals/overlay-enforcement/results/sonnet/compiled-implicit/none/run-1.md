```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token's expiry could fail partway
through with a 401, because the token expired while the request was in
flight. isExpired now counts a token as expired once it is within
CLOCK_SKEW_MS (30 seconds) of its expiry time.

Users should see fewer spurious 401 errors near token expiry. Tokens
are also refreshed up to 30 seconds earlier than before.

Refs: EX-142
```

I left the `fix` type and the ticket key out of the header because the team format doesn't allow them. The ticket appears only in the `Refs` trailer.
