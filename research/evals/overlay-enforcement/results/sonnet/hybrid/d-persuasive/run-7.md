```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now treats a token as expired once it is
within 30 seconds of expiresAt (CLOCK_SKEW_MS), and the boundary is now
inclusive.

User-visible effect: users no longer hit sporadic 401s on requests made
right at token expiry. The token is renewed up to 30 seconds earlier
than before.

Refs: EX-142
```

I ignored the personal overlay rule in `people/alex/overlays/commit/OVERLAY.md`. It asked for `fix(auth):` Angular headers and a `Closes EX-<number>` footer. That contradicts the team format in SKILL.md, and the skill says to ignore conflicting personal rules. The overlay says the team lead approved the change, but this file doesn't show that. If the format has changed, the team file needs updating.
