package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/chuckha/renthelper/avalon"
	"github.com/chuckha/renthelper/slack"

	"github.com/aws/aws-lambda-go/lambda"
)

const (
	configFileFlag     = "config"
	avalonUsernameEnv  = "AVALON_USERNAME"
	avalonPasswordEnv  = "AVALON_PASSWORD"
	rentersEnv         = "RENTERS"
	rentAmountsEnv     = "RENT_AMOUNTS"
	slackChannelIDEnv  = "SLACK_CHANNEL_ID"
	slackOauthTokenEnv = "SLACK_OAUTH_TOKEN"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest() (string, error) {
	cfg := &Config{
		AvalonUsername:  os.Getenv(avalonUsernameEnv),
		AvalonPassword:  os.Getenv(avalonPasswordEnv),
		Renters:         os.Getenv(rentersEnv),
		RentAmounts:     os.Getenv(rentAmountsEnv),
		SlackChannelID:  os.Getenv(slackChannelIDEnv),
		SlackOauthToken: os.Getenv(slackOauthTokenEnv),
	}

	errs := cfg.valid()
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	// Initial data set up.
	ts, err := avalon.GetTenants(cfg.Renters, cfg.RentAmounts)
	if err != nil {
		panic(err)
	}

	// Create the scraper.
	ac, err := avalon.NewClient(cfg.AvalonUsername, cfg.AvalonPassword)
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

	// Print to stdout if we can't send the message to slack.
	if cfg.SlackChannelID == "" || cfg.SlackOauthToken == "" {
		fmt.Println(message)
		return "", nil
	}

	// Create the slack client and post the message.
	sc := slack.NewClient(cfg.SlackOauthToken)
	sc.Post(cfg.SlackChannelID, message)
	return "", nil
}

// requiredStringFlag ensures that the var has a non empty value.
func requiredStringFlag(name, value string) {
	if value == "" {
		fmt.Printf("%v is a required flag. A value must be set.\n", name)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

// Config defines the values needed to run this application
type Config struct {
	AvalonUsername  string `json:"avalon_username"`
	AvalonPassword  string `json:"avalon_password"`
	Renters         string `json:"renters"`
	RentAmounts     string `json:"rent_amounts"`
	SlackChannelID  string `json:"slack_channel_id"`
	SlackOauthToken string `json:"slack_oauth_token"`
}

func (c *Config) valid() []error {
	errlist := make([]error, 0)
	if c.AvalonPassword == "" {
		errlist = append(errlist, fmt.Errorf("avalon password must be set"))
	}
	if c.AvalonUsername == "" {
		errlist = append(errlist, fmt.Errorf("avalon username must be set"))
	}
	if c.RentAmounts == "" {
		errlist = append(errlist, fmt.Errorf("rent amounts must be set (example: 800,700,1000)"))
	}
	if c.Renters == "" {
		errlist = append(errlist, fmt.Errorf("renters must be set (example: person1,person2,person3)"))
	}
	return errlist
}
