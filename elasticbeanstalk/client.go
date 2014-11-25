package elasticbeanstalk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	if err != nil {
		return err
	}
	r.Header.Set("accept", "application/json")
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
		if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
			return err
		}
	}

	return nil
}

// Time is a time.Time whose JSON representation is its floating point
// milliseconds since the epoch.
type Time struct{ time.Time }

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strings.Replace(fmt.Sprintf("%.9E", float64(time.Duration(t.UnixNano())/time.Millisecond)), "E+12", "E9", -1)), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var sec float64
	if err := json.Unmarshal(b, &sec); err != nil {
		return err
	}
	*t = Time{time.Unix(int64(sec), int64(1000*(sec-float64(int64(sec))))).UTC()}
	return nil
}
