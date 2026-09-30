```
auth: Expire tokens 30s early to allow for clock skew

A request that started just before a token expired could fail halfway
through with a 401, because the token was only treated as expired once
its expiry time had passed. Treat a token as expired CLOCK_SKEW_MS (30s)
before expiresAt, and count the exact boundary as expired, so callers
refresh before sending rather than during the request.

User-visible effect: requests made just before expiry no longer fail
with a spurious 401. Tokens are refreshed up to 30s sooner than before.

Refs: EX-142
```

I ignored two rules from alex's personal overlay because they contradict the team's `commit-format` override. The first was the `[EX-142]` prefix in the header, since the team format allows no ticket key there. The second was dropping the `Refs:` trailer, which the team requires. I also left out the Angular `fix(auth):` header, because the team override replaces it. The two attribution lines after `Refs:` come from the session's commit-attribution setting.
