```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before its expiresAt, and a token exactly at the
boundary counts as expired.

User-visible effect: fewer spurious 401s on requests that straddle
expiry. Tokens are refreshed up to 30 seconds earlier than before.

Refs: EX-142
```

The skill has two conflicting formats. I used the team format (area prefix, `Refs:` trailer) and left out the Angular-style `fix(auth):` header and `Closes` footer from the "Personal additions (alex)" section. That section is alex's, not yours, and I can't verify the team lead's approval it cites. If the switch is real, tell me and I'll redo the message in that format.
