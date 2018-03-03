# Rent Helper

This is a script that logs into [Avalon][avalon]'s rent portal, grabs how much
is owed, splits the configured rent amounts amongst the tenants and divides
the remaining amount (presumably the utilities) evenly across each tenant.

It then posts to slack or standard out.

## How to run?

### Installation

Install with `go get github.com/chuckha/cmd/renthelper`

### Configuration

```json
{
    "avalon_username": "avalon_login@example.com",
    "avalon_password": "avalon_password",
    "renters": "p1,p2,p3",
    "rent_amounts": "200,300,400",
    "slack_channel_id": "S0M3CH4N",
    "slack_oauth_token": "oauthtoken"
}
```

If you omit the slack id/token renthelper will print to stdout.

#### Configuration directories

Put configuration files in:

- `/etc/renthelpers/*.json`
- `$HOME/.renthelpers/*.json`

and `renthelper` will load them automatically.

#### Specify a custom configuration

Create a configuration file and pass it in with the `-config` flag.

### Run with a custom configuration

```
renthelper -config config.json
```

##

[avalon]: https://www.avaloncommunities.com/
