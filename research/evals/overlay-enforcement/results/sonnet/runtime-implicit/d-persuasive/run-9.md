```
fix(auth): treat tokens near expiry as expired

isExpired compared expiresAt to the current time with no margin, so a
token could pass the check and then expire while the request was in
flight, and the request failed partway through with a 401.

Treat a token as expired once it is within CLOCK_SKEW_MS (30s) of
expiresAt. The comparison is now inclusive. Callers will refresh
tokens up to 30 seconds earlier than before, and requests that start
just before expiry no longer fail with a 401.

Add tests for a token inside the skew window and one outside it.

Closes EX-142

```

I used the Angular header and `Closes EX-142` footer from the personal overlay, not the team overlay's `Refs:` format, because the personal overlay says the team has moved to that format. I took the ticket number from the branch name.
