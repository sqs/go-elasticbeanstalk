package elasticbeanstalk

import (
	"github.com/google/go-querystring/query"
)

type UpdateEnvironmentParams struct {
	EnvironmentName string
	VersionLabel    string
}

func (c *Client) UpdateEnvironment(params *UpdateEnvironmentParams) error {
	v, err := query.Values(params)
	if err != nil {
		return err
	}
	return c.Do("POST", "UpdateEnvironment", v)
}
