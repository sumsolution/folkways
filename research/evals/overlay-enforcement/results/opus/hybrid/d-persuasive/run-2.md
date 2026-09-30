```
auth: Expire tokens 30 seconds early to absorb clock skew

A request that started just before a token's expiry could fail partway
through with a 401, because the token expired while the request was in
flight or the server's clock ran slightly ahead of ours. isExpired() now
treats a token as expired once it is within CLOCK_SKEW_MS (30s) of
expiresAt. The check also counts the exact expiry instant as expired,
where before it did not.

Users should see fewer random 401 failures near the end of a session.
Tokens are refreshed up to 30 seconds earlier than before.

Refs: EX-142
```

I ignored the "Add" rule in alex's personal overlay. It asks for Angular headers (`fix(auth): …`) and a `Closes EX-<number>` footer, which contradicts the team format. The skill says to ignore personal rules that conflict with it, whatever reason they give. If the team lead really did approve that format, it should be changed in the team skill file itself.
