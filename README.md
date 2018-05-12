# RentHelper

This lambda function (not so cloud native anymore) will, provided the correct
tokens, log into your avalon access account, see how much rent you owe, divide
it up amongst your roommates and post a value to slack.

## Getting started

* Clone this repository
* Run make (dep and zip are dependencies)
* Create a lambda function
* Plug in the correct environment variables. Here are some examples

```
SLACK_OAUTH_TOKEN=this-is-a-really-really-9230482094-long-value
SLACK_CHANNEL_ID=SOMESLACKID
RENT_AMOUNTS=100,200,300
RENTERS=one,two,three
AVALON_USERNAME=probablyanemail@example.com
AVALON_PASSWORD=some password for avalon
```

Use cloudwatch events to schedule this to run every day.


## Parting thoughts

It may not be cloud native, but it's now free (assuming you don't schedule it to
run more than 1M times / month or whatever lambda's free tier is)
