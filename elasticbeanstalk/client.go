package elasticbeanstalk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/crowdmob/goamz/aws"
)

type Client struct {
	BaseURL    *url.URL
	Auth       aws.Auth
	Region     aws.Region
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

func (c *Client) Do(method string, operation string, params url.Values, respData interface{}) error {
	url := c.BaseURL.ResolveReference(&url.URL{RawQuery: fmt.Sprintf("Operation=%s&%s", operation, params.Encode())})
	r, err := http.NewRequest(method, url.String(), nil)
	r.Header.Set("X-Amz-Date", time.Now().UTC().Format(aws.ISO8601BasicFormat))
	signer := aws.NewV4Signer(c.Auth, "elasticbeanstalk", c.Region)
	signer.Sign(r)

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("http status code %d (%s): %s", resp.StatusCode, http.StatusText(resp.StatusCode), msg)
	}

	if respData != nil {
		err = json.NewDecoder(resp.Body).Decode(respData)
		if err != nil {
			return err
		}
	}

	return nil
}
