package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/chuckha/renthelper/avalon"
	"github.com/chuckha/renthelper/slack"
)

const (
	configFileFlag = "config"
)

var (
	globalConfigPath = path.Join("/", "etc", "renthelper")
	userConfigPath   = path.Join(os.Getenv("HOME"), ".renthelper")
)

func main() {
	configFile := flag.String(configFileFlag, "", "path to a config file")
	flag.Parse()
	cfg, err := loadConfig(*configFile)
	if err != nil {
		fmt.Println("error loading config", err)
		os.Exit(1)
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

	// Print to stdout if we can't send the mnessage to slack.
	if cfg.SlackChannelID == "" || cfg.SlackOauthToken == "" {
		fmt.Println(message)
		return
	}

	// Create the slack client and post the message.
	sc := slack.NewClient(cfg.SlackOauthToken)
	sc.Post(cfg.SlackChannelID, message)
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

func loadConfig(explicit string) (*Config, error) {
	var finalConfig Config

	// Quick exit if we've defined an explicit conifg
	if explicit != "" {
		data, err := ioutil.ReadFile(explicit)
		if err != nil {
			return nil, fmt.Errorf("error with explicit config file: %v", err)
		}

		return &finalConfig, json.Unmarshal(data, &finalConfig)
	}

	searchPaths := []string{globalConfigPath, userConfigPath}

	// Go through each config path and load a config file. Each subsequent config will override the previous one.
	for _, path := range searchPaths {
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			// bail if we're in an error state
			if err != nil {
				return err
			}

			// If it's not a .json file we don't care about it
			if !strings.HasSuffix(info.Name(), ".json") {
				return nil
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			return json.Unmarshal(data, &finalConfig)
		}); err != nil {
			return nil, err
		}
	}
	return &finalConfig, nil
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
