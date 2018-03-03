package avalon

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL = "https://www.avalonaccess.com/"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	Host string
	Doer
}

func NewClient(username, password string) (*Client, error) {
	client, err := clientWithCookieJar()
	if err != nil {
		return nil, fmt.Errorf("failed to get http client")
	}
	c := &Client{
		Host: baseURL,
		Doer: client,
	}
	err = c.Login(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	return c, nil
}

// login sets the cookies on the avalon client
func (c *Client) Login(username, password string) error {
	loginURL := fmt.Sprintf("%v%v", c.Host, "UserProfile/LogOn")
	data := url.Values{}
	data.Add("UserName", username)
	data.Add("password", password)
	r, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("could not get new request: %v", err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(r)
	if err != nil {
		return fmt.Errorf("failed to get response from POST: %v", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("bad response: %v, [%v]", r.URL.String(), resp.StatusCode)
	}
	return nil
}

// Does some regex parsing out of html to get the balance owed.
func (c *Client) GetBalance() (*BalanceSheet, error) {
	url := fmt.Sprintf("%v%v", c.Host, totalOwedURL())
	ledgerRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make ledger request: %v", err)
	}
	ledgerResponse, err := c.Do(ledgerRequest)
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
