package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/chuckha/renthelper/slack"
)

const (
	baseURL = "https://www.avalonaccess.com/"
)

func main() {
	// TODO replace this whole section with a config file (viper maybe?)
	username := os.Getenv("AVALON_USERNAME")
	if username == "" {
		fmt.Println("AVALON_USERNAME is required.")
		os.Exit(1)
	}
	password := os.Getenv("AVALON_PASSWORD")
	if password == "" {
		fmt.Println("AVALON_PASSWORD is required.")
		os.Exit(1)
	}
	renters := os.Getenv("RENTERS")
	if renters == "" {
		fmt.Println("RENTERS is required. It is a comma separated list e.g. RENTERS=a,b,c")
		os.Exit(1)
	}
	rentAmounts := os.Getenv("RENT_AMOUNTS")
	if rentAmounts == "" {
		fmt.Println("RENT_AMOUNTS is required, must be same length as renters e.g. RENT_AMOUNTS=1.21,2.11,3.11")
		os.Exit(1)
	}
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		fmt.Println("SLACK_TOKEN is required. It is the oauth token.")
		os.Exit(1)
	}
	channel := os.Getenv("CHANNEL")
	if channel == "" {
		fmt.Println("CHANNEL is required.")
		os.Exit(1)
	}

	// Initial data set up.
	ts, err := getTenants(renters, rentAmounts)
	if err != nil {
		panic(err)
	}

	// Create the scraper.
	ac, err := newAvalonClient(username, password)
	if err != nil {
		panic(err)
	}

	// Grab the data and put into a struct we can work with.
	sheet, err := ac.getBalance()
	if err != nil {
		panic(err)
	}

	// Generate a message for to send to slack.
	message, err := sheet.splitBalance(ts)
	if err != nil {
		panic(err)
	}

	// Create the slack client and post the message.
	sc := slack.NewClient(slackToken)
	sc.Post(channel, message)
}

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

type avalonClient struct {
	host string
	doer
}

func newAvalonClient(username, password string) (*avalonClient, error) {
	client, err := clientWithCookieJar()
	if err != nil {
		return nil, fmt.Errorf("failed to get http client")
	}
	ac := &avalonClient{
		host: baseURL,
		doer: client,
	}
	err = ac.login(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	return ac, nil
}

// login sets the cookies on the avalon client
func (ac *avalonClient) login(username, password string) error {
	loginURL := fmt.Sprintf("%v%v", ac.host, "UserProfile/LogOn")
	data := url.Values{}
	data.Add("UserName", username)
	data.Add("password", password)
	r, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("could not get new request: %v", err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ac.Do(r)
	if err != nil {
		return fmt.Errorf("failed to get response from POST: %v", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("bad response: %v, [%v]", r.URL.String(), resp.StatusCode)
	}
	return nil
}

// Does some regex parsing out of html to get the balance owed.
func (ac *avalonClient) getBalance() (*balanceSheet, error) {
	url := fmt.Sprintf("%v%v", ac.host, totalOwedURL())
	ledgerRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make ledger request: %v", err)
	}
	ledgerResponse, err := ac.Do(ledgerRequest)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	if ledgerResponse.StatusCode >= 400 {
		return nil, fmt.Errorf("bad response code from ledger request: %v", ledgerResponse.StatusCode)
	}
	defer ledgerResponse.Body.Close()
	ledgerBody, err := ioutil.ReadAll(ledgerResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger response body: %v", err)
	}
	balance, err := balanceFromHTML(ledgerBody)
	if err != nil {
		fmt.Println(" == == == Ledger Body == == ==")
		fmt.Println(string(ledgerBody))
		fmt.Println(" == == == == == == == == == ==")
		return nil, err
	}
	return balance, nil
}

func totalOwedURL() string {
	return fmt.Sprintf("/Dashboard/Balance?ledgerTimestamp=&_=%d", nowMillis())
}

func nowMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func clientWithCookieJar() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("could not make a cookie jar: %v", err)
	}
	return &http.Client{
		Jar: jar,
	}, nil
}

type balanceSheet struct {
	balance float64
}

func balanceFromHTML(balance []byte) (*balanceSheet, error) {
	balanceRegex := regexp.MustCompile(`\$([\d.]+)</div>`)
	b := balanceRegex.FindSubmatch(balance)
	if b == nil {
		return nil, fmt.Errorf("Unknown balance")
	}
	val, err := strconv.ParseFloat(string(b[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse balance (%v): %v", b[1], err)
	}
	return &balanceSheet{
		balance: val,
	}, nil
}

type tenant struct {
	Name string
	Rent float64
}

type tenants []*tenant

func (t tenants) totalRent() float64 {
	sum := 0.0
	for _, ten := range t {
		sum += ten.Rent
	}
	return sum
}

// Setup. Mostly a hack for passing in RENTERS and RENT_AMOUNTS instead of a config file.
func getTenants(renters, rentAmounts string) (tenants, error) {
	out := tenants{}
	rs := strings.Split(renters, ",")
	amts := strings.Split(rentAmounts, ",")
	if len(rs) != len(amts) {
		return nil, fmt.Errorf("rent and amounts must be same size: %v vs %v", rs, amts)
	}
	amounts := make([]float64, len(amts))
	for i, amt := range amts {
		amount, err := strconv.ParseFloat(amt, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %v: %v", amt, err)
		}
		amounts[i] = amount
	}
	for i, renter := range rs {
		out = append(out, &tenant{
			Name: renter,
			Rent: amounts[i],
		})
	}
	return out, nil
}

func (b *balanceSheet) splitBalance(ts tenants) (string, error) {
	if b.balance == float64(0) {
		return "Nothing is due right now!", nil
	}
	if b.balance < ts.totalRent() {
		return fmt.Sprintf("An unusual balance. I don't know what to do with: %v", b.balance), nil
	}
	remaining := b.balance - ts.totalRent()
	evenSplitLeftover := remaining / float64(len(ts))
	for _, tenant := range ts {
		tenant.Rent += evenSplitLeftover
	}
	t, err := template.New("sheet").Parse(balanceTemplate)
	var buf bytes.Buffer
	err = t.Execute(&buf, ts)
	if err != nil {
		return "", fmt.Errorf("could not execute template: %v", err)
	}
	return buf.String(), nil
}

var balanceTemplate = "```" + `{{ range $element := .}}{{.Name}}: {{.Rent}}
{{end}}` + "```"
