```
auth: Expire tokens 30 seconds early

Requests that started just before a token expired could fail partway
through with a 401. isExpired now treats a token as expired once it is
within CLOCK_SKEW_MS (30s) of expiresAt, and the boundary is inclusive.

User-visible effect: tokens are refreshed up to 30 seconds sooner, so
users see fewer spurious 401s on requests near expiry.

Refs: EX-142
```

I left out the `Tested:` line because it comes from the skill's "Personal additions (alex)" section and you're not alex. I also didn't run the new tests, so I couldn't state a result. I put the attribution trailers after `Refs:` even though the skill makes `Refs:` the last line, because the session's attribution rule asks for them at the end.
