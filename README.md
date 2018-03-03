# Rent Helper

This is a script that logs into [Avalon][avalon]'s rent portal, grabs how much
is owed, splits the configured rent amounts amongst the tenants and divides
the remaining amount (presumably the utilities) evenly across each tenant.

It then posts to slack or standard out.

## How to run?

### Installation

Install with `go get github.com/chuckha/cmd/renthelper`

### Run

Run with a command like this:

```
renthelper \
-avalon-username=emailaddress@example.com \
-avalon-password=somesillypassword \
-renters=person1,person2,person3 \
-rent-amounts=1000,1100,1200 \
-slack-channel-id=SLACKCHANNELID \
-slack-token=some-slack-oauth-token
```

[avalon]: https://www.avaloncommunities.com/
