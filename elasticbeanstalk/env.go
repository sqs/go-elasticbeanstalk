package elasticbeanstalk

import (
	"fmt"
	"net/url"

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
	DateCreated       Time
	DateUpdated       Time
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
		DescribeEnvironmentsResponse struct {
			DescribeEnvironmentsResult struct {
				Environments []*EnvironmentDescription
			}
		}
	}
	err = c.Do("GET", "DescribeEnvironments", v, &o)
	return o.DescribeEnvironmentsResponse.DescribeEnvironmentsResult.Environments, err
}

// A ConfigurationSettingsDescription describes the settings for a
// configuration.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/APIReference/API_ConfigurationSettingsDescription.html.
type ConfigurationSettingsDescription struct {
	ApplicationName   string
	DateCreated       Time
	DateUpdated       Time
	DeploymentStatus  string
	Description       string `json:",omitempty"`
	EnvironmentName   string
	OptionSettings    ConfigurationOptionSettings
	SolutionStackName string
	TemplateName      string `json:",omitempty"`
}

// A ConfigurationSettings is a list of
// ConfigurationSettingsDescription that provides easy access to
// combined configuration settings.
type ConfigurationSettings []*ConfigurationSettingsDescription

// Environ returns a map of all environment variables set in the
// configuration settings.
func (s ConfigurationSettings) Environ() map[string]string {
	m := map[string]string{}
	for _, csd := range s {
		m0 := csd.OptionSettings.Environ()
		for k, v := range m0 {
			m[k] = v
		}
	}
	return m
}

// DescribeConfigurationSettingsParams specifies parameters for a
// DescribeConfigurationSettings request.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/APIReference/API_DescribeConfigurationSettings.html.
type DescribeConfigurationSettingsParams struct {
	ApplicationName string
	EnvironmentName string `url:",omitempty"`
	TemplateName    string `url:",omitempty"`
}

func (c *Client) DescribeConfigurationSettings(params *DescribeConfigurationSettingsParams) (ConfigurationSettings, error) {
	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	var o struct {
		DescribeConfigurationSettingsResponse struct {
			DescribeConfigurationSettingsResult struct {
				ConfigurationSettings []*ConfigurationSettingsDescription
			}
		}
	}
	err = c.Do("GET", "DescribeConfigurationSettings", v, &o)
	return o.DescribeConfigurationSettingsResponse.DescribeConfigurationSettingsResult.ConfigurationSettings, err
}

// A ConfigurationOptionSettings is a list of
// ConfigurationOptionSetting that provides easy access to environment
// variables specified within.
type ConfigurationOptionSettings []ConfigurationOptionSetting

// Environ returns a map of all environment variables set in the
// option settings.
func (opts ConfigurationOptionSettings) Environ() map[string]string {
	m := map[string]string{}
	for _, opt := range opts {
		if opt.Namespace == envVarNamespace {
			m[opt.OptionName] = opt.Value
		}
	}
	return m
}

type UpdateEnvironmentParams struct {
	EnvironmentName string
	VersionLabel    string `url:",omitempty"`

	OptionSettings ConfigurationOptionSettings `url:"-"`
}

const envVarNamespace = "aws:elasticbeanstalk:application:environment"

// AddEnv adds the specified environment variable name and value to
// OptionSettings.
func (p *UpdateEnvironmentParams) AddEnv(name, value string) {
	p.OptionSettings = append(p.OptionSettings, ConfigurationOptionSetting{
		Namespace:  envVarNamespace,
		OptionName: name,
		Value:      value,
	})
}

// optionSettingsValues returns a url.Values for the
// (UpdateEnvironmentParams).OptionSettings field entries. Each entry yields 3
// keys whose names are prefixed with `OptionSettings.member.N.`.
func (p *UpdateEnvironmentParams) optionSettingsValues() url.Values {
	if len(p.OptionSettings) == 0 {
		return nil
	}
	v := make(url.Values)
	for i, s := range p.OptionSettings {
		kp := fmt.Sprintf("OptionSettings.member.%d", i+1)
		v.Set(kp+".Namespace", s.Namespace)
		v.Set(kp+".OptionName", s.OptionName)
		v.Set(kp+".Value", s.Value)
	}
	return v
}

// ConfigurationOptionSetting is a specification identifying an individual
// configuration option along with its current value.
//
// See
// http://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_ConfigurationOptionSetting.html.
type ConfigurationOptionSetting struct {
	Namespace  string
	OptionName string
	Value      string
}

func (c *Client) UpdateEnvironment(params *UpdateEnvironmentParams) error {
	v, err := query.Values(params)
	if err != nil {
		return err
	}

	osv := params.optionSettingsValues()
	for k, vs := range osv {
		v[k] = vs
	}

	return c.Do("POST", "UpdateEnvironment", v, nil)
}
