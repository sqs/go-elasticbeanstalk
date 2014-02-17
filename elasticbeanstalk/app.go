package elasticbeanstalk

import (
	"github.com/google/go-querystring/query"
)

type CreateApplicationVersionParams struct {
	ApplicationName      string
	VersionLabel         string
	Description          string
	SourceBundleS3Bucket string `url:"SourceBundle.S3Bucket"`
	SourceBundleS3Key    string `url:"SourceBundle.S3Key"`
}

func (c *Client) CreateApplicationVersion(params *CreateApplicationVersionParams) error {
	// AWS wants "Description=", not just "Description", if empty, so force it
	// to be non-empty TODO(sqs):try omitempty
	if params.Description == "" {
		params.Description = "_"
	}
	v, err := query.Values(params)
	if err != nil {
		return err
	}
	return c.Do("POST", "CreateApplicationVersion", v)
}
