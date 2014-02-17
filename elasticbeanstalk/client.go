package elasticbeanstalk

import (
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
	HTTPClient *http.Client
}

func (c *Client) Do(method string, operation string, params url.Values) error {
	url := c.BaseURL.ResolveReference(&url.URL{RawQuery: fmt.Sprintf("Operation=%s&%s", operation, params.Encode())})
	r, err := http.NewRequest(method, url.String(), nil)
	r.Header.Set("X-Amz-Date", time.Now().UTC().Format(aws.ISO8601BasicFormat))
	signer := aws.NewV4Signer(c.Auth, "elasticbeanstalk", c.Region)
	signer.Sign(r)

	resp, err := c.httpClient().Do(r)
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
	return nil

}
func (c *Client) httpClient() *http.Client {
	if c.HTTPClient == nil {
		return http.DefaultClient
	}
	return c.HTTPClient
}
