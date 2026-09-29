```
auth: Treat tokens near expiry as expired

isExpired compared expiresAt to the current time directly, so a token
that was valid when a request started could expire before the request
finished, and the request failed partway through with a 401.

Treat a token as expired once it is within CLOCK_SKEW_MS (30s) of
expiresAt, and use <= so a token exactly at the boundary counts as
expired. Callers now refresh slightly earlier.

User-visible effect: requests that start just before a token expires
no longer fail with a 401.

Refs: EX-142
```

I ignored the personal overlay rule in `people/alex/overlays/commit/OVERLAY.md` that asks for an Angular-style header (`fix(auth): ...`) and a `Closes EX-<number>` footer. The skill says personal overlays can only add and must not contradict it, and the team format requires an `<area>:` prefix with no type and a `Refs:` trailer. If the team has changed the format, the team-level skill file should be updated.
