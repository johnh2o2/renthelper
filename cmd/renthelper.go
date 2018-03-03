package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/chuckha/renthelper/avalon"
	"github.com/chuckha/renthelper/slack"
)

const (
	avalonUsernameFlag  = "avalon-username"
	avalonPasswordFlag  = "avalon-password"
	rentersFlag         = "renters"
	rentAmountsFlag     = "rent-amounts"
	slackOauthTokenFlag = "slack-oauth-token"
	slackChannelFlag    = "slack-channel-id"
)

func main() {
	avalonUsername := flag.String(avalonUsernameFlag, "", "username for avalon communities")
	avalonPassword := flag.String(avalonPasswordFlag, "", "password for avalon communities")
	// TODO(chuckha) use a custom flag type here
	renters := flag.String(rentersFlag, "", "a comma separated list of renters")
	rentAmounts := flag.String(rentAmountsFlag, "", "a comma separated list of rent amounts that corresponds to the renters")
	slackOauthToken := flag.String(slackOauthTokenFlag, "", "the slack oauth token")
	slackChannel := flag.String(slackChannelFlag, "", "the slack channel ID to post the result to")
	flag.Parse()

	// Flag validation
	requiredStringFlag(avalonUsernameFlag, *avalonUsername)
	requiredStringFlag(avalonPasswordFlag, *avalonPassword)
	requiredStringFlag(rentersFlag, *renters)
	requiredStringFlag(rentAmountsFlag, *rentAmounts)

	// Initial data set up.
	ts, err := avalon.GetTenants(*renters, *rentAmounts)
	if err != nil {
		panic(err)
	}

	// Create the scraper.
	ac, err := avalon.NewClient(*avalonUsername, *avalonPassword)
	if err != nil {
		panic(err)
	}

	// Grab the data and put into a struct we can work with.
	sheet, err := ac.GetBalance()
	if err != nil {
		panic(err)
	}

	// Generate a message for to send to slack.
	message, err := sheet.SplitBalance(ts)
	if err != nil {
		panic(err)
	}

	// Print to stdout if we can't send the mnessage to slack.
	if *slackChannel == "" || *slackOauthToken == "" {
		fmt.Println(message)
		return
	}

	// Create the slack client and post the message.
	sc := slack.NewClient(*slackOauthToken)
	sc.Post(*slackChannel, message)
}

// requiredStringFlag ensures that the var has a non empty value.
func requiredStringFlag(name, value string) {
	if value == "" {
		fmt.Printf("%v is a required flag. A value must be set.\n", name)
		flag.PrintDefaults()
		os.Exit(1)
	}
}
