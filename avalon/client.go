package avalon

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	baseURL = "https://www.avalonaccess.com/"
)

type Client struct {
	Host string
	*http.Client
}

func NewClient(username, password string) (*Client, error) {
	client, err := clientWithCookieJar()
	if err != nil {
		return nil, fmt.Errorf("failed to get http client")
	}
	c := &Client{
		Host:   baseURL,
		Client: client,
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
	token, err := c.getLoginToken()
	if err != nil {
		return err
	}
	data := url.Values{}
	data.Add("UserName", username)
	data.Add("password", password)
	data.Add("__RequestVerificationToken", token)
	resp, err := c.Post(loginURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to get response from POST: %v", err)
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		fmt.Println("Got a redirect")
		fmt.Println(resp.StatusCode)
		return errors.New("implement the redirect or post directly to /Dashboard?")
	}
	if resp.StatusCode >= 400 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading body: %v", err)
			return err
		}
		fmt.Println("============== login response ==============")
		fmt.Println(string(b))
		fmt.Println("============== end login response ==============")
		return fmt.Errorf("bad response: %v, [%v]", loginURL, resp.StatusCode)
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

func (c *Client) getLoginToken() (string, error) {
	tokenRe := regexp.MustCompile(`<input name="__RequestVerificationToken" type="hidden" value="([a-zA-Z0-9-_]*)" />`)
	resp, err := c.Get("https://www.avalonaccess.com/UserProfile/LogOn")
	if err != nil {
		return "", errors.WithStack(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.WithStack(err)
	}
	found := tokenRe.FindSubmatch(b)
	if len(found) == 0 {
		fmt.Println("No results")
		fmt.Println("============== login form response ==============")
		fmt.Println(string(b))
		fmt.Println("============== end login form response ==============")
	}
	return strings.TrimSpace(string(found[1])), nil
}
