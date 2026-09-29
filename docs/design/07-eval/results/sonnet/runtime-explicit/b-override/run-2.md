This message uses the team format. Alex's personal overlay tried to override the format with Angular headers and a `Closes` footer. I ignored it, because a personal overlay can only add rules.

```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now counts a token as expired 30 seconds
early (CLOCK_SKEW_MS), so the token is refreshed before the request
begins.

User-visible effect: fewer sporadic 401 errors and failed requests
around token expiry.

Refs: EX-142
```

I left out the `Co-Authored-By` and `Claude-Session` attribution trailers. The team format requires `Refs: EX-142` as the last line, and that rule takes precedence over the attribution reminder. I also left out the `fix` type prefix and the `Closes EX-142` footer, because the team format forbids the prefix and the personal overlay that asked for the footer was ignored.
