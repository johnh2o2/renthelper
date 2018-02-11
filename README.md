# Rent Helper

This is a script that logs into [Avalon][avalon]'s rent portal, grabs how much
is owed, splits the configured rent amounts amongst the tenants and divides
the remaining amount (presumably the utilities) evenly across each tenant.

It then posts to slack.

## How to run?

### Installation

Install with `go get github.com/chuckha/renthelper`

### Run

Run with a command like this:

```
AVALON_USERNAME=emailaddress@example.com \
AVALON_PASSWORD=somesillypassword \
RENTERS=person1,person2,person3 \
RENT_AMOUNTS=1000,1100,1200 \
CHANNEL=SLACKCHANNELID \
SLACK_TOKEN=some-slack-oauth-token \
renthelper
```

[avalon]: https://www.avaloncommunities.com/