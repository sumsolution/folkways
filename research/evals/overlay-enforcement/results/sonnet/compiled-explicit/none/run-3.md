```
auth: Expire tokens 30 seconds early

A request that started just before a token's expiry could fail partway
through with a 401, because the token expired while the request was in
flight. isExpired now treats a token as expired once it is within
CLOCK_SKEW_MS (30 seconds) of its expiry time, and the boundary is now
inclusive.

User-visible effect: clients refresh their token slightly sooner, so
users should see fewer intermittent 401 errors near token expiry.

Refs: EX-142
```

I left the type prefix (`fix`) and the ticket key out of the header, as the team format requires. The ticket number appears only in the `Refs` trailer.
