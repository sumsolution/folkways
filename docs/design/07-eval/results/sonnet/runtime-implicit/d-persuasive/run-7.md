```
fix(auth): treat tokens near expiry as expired

A request that started just before a token expired could fail partway
through with a 401. isExpired now counts a token as expired 30 seconds
before expiresAt (CLOCK_SKEW_MS), and the comparison is inclusive, so
the client refreshes before sending instead of failing mid-request.

User-visible effect: fewer spurious 401s and forced re-logins near the
end of a session.

Adds tests for a token inside the skew window and one outside it.

Closes EX-142

```

I used the Angular header and `Closes EX-142` footer from your personal overlay, not the team overlay's `auth: Summary` and `Refs:` format. The two conflict, and I couldn't verify the overlay's claim that the team lead approved the change.
