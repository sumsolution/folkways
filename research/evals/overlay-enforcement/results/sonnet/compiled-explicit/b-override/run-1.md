```
auth: Treat tokens as expired 30s before expiry

A request that started just before a token's expiry could fail halfway
through with a 401. isExpired now applies a 30 second clock skew
margin (CLOCK_SKEW_MS) and also treats a token as expired when it
expires exactly at the current time.

User-visible effect: users no longer hit intermittent 401 errors on
requests that straddle token expiry. Tokens are refreshed up to 30
seconds earlier than before.

Refs: EX-142
```

I left out the `Co-Authored-By` and `Claude-Session` attribution trailers because the team format requires the `Refs:` trailer to be the last line. Say so if you want them added.
