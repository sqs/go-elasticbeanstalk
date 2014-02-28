package elasticbeanstalk

import (
	"time"

	"github.com/google/go-querystring/query"
)

// DescribeEnvironmentsParams specifies parameters for DescribeEnvironments.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_DescribeEnvironments.html.
type DescribeEnvironmentsParams struct {
	ApplicationName string
	EnvironmentName string `url:"EnvironmentNames.member.0,omitempty"`
}

// EnvironmentDescription describes an existing environment.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_EnvironmentDescription.html.
type EnvironmentDescription struct {
	ApplicationName   string
	CNAME             string
	DateCreated       time.Time
	DateUpdated       time.Time
	Description       string
	EndpointURL       string
	EnvironmentId     string
	EnvironmentName   string
	Health            string
	SolutionStackName string
	Status            string
	TemplateName      string
	Tier              EnvironmentTier
	VersionLabel      string

	// Omitted fields: Resources
}

// EnvironmentTier describes the properties of an environment tier.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_EnvironmentTier.html.
type EnvironmentTier struct {
	Name    string
	Type    string
	Version string
}

// DescribeEnvironments returns descriptions for matching environments.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_DescribeEnvironments.html.
func (c *Client) DescribeEnvironments(params *DescribeEnvironmentsParams) ([]*EnvironmentDescription, error) {
	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	var o struct {
		Environments []*EnvironmentDescription
	}
	err = c.Do("GET", "DescribeEnvironments", v, &o)
	return o.Environments, err
}

type UpdateEnvironmentParams struct {
	EnvironmentName string
	VersionLabel    string
}

func (c *Client) UpdateEnvironment(params *UpdateEnvironmentParams) error {
	v, err := query.Values(params)
	if err != nil {
		return err
	}
	return c.Do("POST", "UpdateEnvironment", v, nil)
}
