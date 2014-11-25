package elasticbeanstalk

import (
	"log"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kr/pretty"
)

func floatTime(t *testing.T, timeStr string) string {
	tm := mustParseTime(t, timeStr).Round(time.Millisecond)
	b, _ := Time{tm}.MarshalJSON()
	return string(b)
}

func TestDescribeEnvironments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		writeJSON(w, `
{
    "DescribeEnvironmentsResponse": {"DescribeEnvironmentsResult": {"Environments": [
        {
            "ApplicationName": "app",
            "CNAME": "app-env.elasticbeanstalk.com",
            "DateCreated": `+floatTime(t, "2014-02-28T00:22:21.474Z")+`,
            "DateUpdated": `+floatTime(t, "2014-02-28T00:33:47.684Z")+`,
            "EndpointURL": "awseb-e-n-AWSEBLoa-MILTONWOOF-1234567.us-west-2.elb.amazonaws.com",
            "EnvironmentId": "e-abcdef1234",
            "EnvironmentName": "app-env",
            "Health": "Green",
            "SolutionStackName": "64bit Amazon Linux 2013.09 running Node.js",
            "Status": "Ready",
            "Tier": {
                "Name": "WebServer",
                "Type": "Standard",
                "Version": "1.0"
            },
            "VersionLabel": "app-123"
        }
    ]
}}}
`)
	})

	want := []*EnvironmentDescription{
		{
			ApplicationName:   "app",
			CNAME:             "app-env.elasticbeanstalk.com",
			DateCreated:       mustParseTime(t, "2014-02-28T00:22:21.474Z"),
			DateUpdated:       mustParseTime(t, "2014-02-28T00:33:47.684Z"),
			EndpointURL:       "awseb-e-n-AWSEBLoa-MILTONWOOF-1234567.us-west-2.elb.amazonaws.com",
			EnvironmentId:     "e-abcdef1234",
			EnvironmentName:   "app-env",
			Health:            "Green",
			SolutionStackName: "64bit Amazon Linux 2013.09 running Node.js",
			Status:            "Ready",
			Tier: EnvironmentTier{
				Name:    "WebServer",
				Type:    "Standard",
				Version: "1.0",
			},
			VersionLabel: "app-123",
		},
	}

	envs, err := client.DescribeEnvironments(&DescribeEnvironmentsParams{})
	if err != nil {
		t.Errorf("DescribeEnvironments returned error: %v", err)
	}

	normTime(&want[0].DateCreated)
	normTime(&want[0].DateUpdated)
	log.Printf("%s != %s", want[0].DateCreated, envs[0].DateCreated)
	if !reflect.DeepEqual(envs, want) {
		t.Errorf("DescribeEnvironments returned %+v, want %+v", asJSON(t, envs), asJSON(t, want))
	}
}

func TestConfigurationSettings_Environ(t *testing.T) {
	got := ConfigurationSettings{
		{
			OptionSettings: ConfigurationOptionSettings{
				{Namespace: "aws:elasticbeanstalk:application:environment", OptionName: "k1", Value: "v1"},
			},
		},
		{
			OptionSettings: ConfigurationOptionSettings{
				{Namespace: "aws:elasticbeanstalk:application:environment", OptionName: "k2", Value: "v2"},
			},
		},
	}.Environ()
	want := map[string]string{"k1": "v1", "k2": "v2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDescribeConfigurationSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		writeJSON(w, `
{
    "DescribeConfigurationSettingsResponse": {"DescribeConfigurationSettingsResult": {"ConfigurationSettings": [
        {
            "ApplicationName": "app",
            "DateCreated": `+floatTime(t, "2014-02-28T00:22:21.474Z")+`,
            "DateUpdated": `+floatTime(t, "2014-02-28T00:33:47.684Z")+`,
            "DeploymentStatus": "deployed",
            "Description": "d",
            "EnvironmentName": "app-env",
            "OptionSettings": [
                {
                    "Namespace": "n",
                    "OptionName": "o",
                    "Value": "v"
                }
            ],
            "SolutionStackName": "64bit Amazon Linux 2013.09 running Node.js",
            "TemplateName": "t"
        }
    ]
}}}
`)
	})

	want := ConfigurationSettings{
		{
			ApplicationName:  "app",
			DateCreated:      mustParseTime(t, "2014-02-28T00:22:21.474Z"),
			DateUpdated:      mustParseTime(t, "2014-02-28T00:33:47.684Z"),
			DeploymentStatus: "deployed",
			Description:      "d",
			EnvironmentName:  "app-env",
			OptionSettings: ConfigurationOptionSettings{
				{Namespace: "n", OptionName: "o", Value: "v"},
			},
			SolutionStackName: "64bit Amazon Linux 2013.09 running Node.js",
			TemplateName:      "t",
		},
	}

	cs, err := client.DescribeConfigurationSettings(&DescribeConfigurationSettingsParams{})
	if err != nil {
		t.Errorf("DescribeConfigurationSettings returned error: %v", err)
	}

	normTime(&want[0].DateCreated)
	normTime(&want[0].DateUpdated)
	if !reflect.DeepEqual(cs, want) {
		t.Errorf("DescribeConfigurationSettings returned %v, want %v", asJSON(t, cs), asJSON(t, want))
	}
}

func TestConfigurationOptionSettings_Environ(t *testing.T) {
	got := ConfigurationOptionSettings{
		{Namespace: "aws:elasticbeanstalk:application:environment", OptionName: "k1", Value: "v1"},
		{Namespace: "aws:elasticbeanstalk:application:environment", OptionName: "k2", Value: "v2"},
	}.Environ()
	want := map[string]string{"k1": "v1", "k2": "v2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUpdateEnvironment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
	})

	err := client.UpdateEnvironment(&UpdateEnvironmentParams{})
	if err != nil {
		t.Errorf("UpdateEnvironment returned error: %v", err)
	}
}

func TestUpdateEnvironment_OptionSettings_Env(t *testing.T) {
	setup()
	defer teardown()

	wantParams := url.Values{
		"Operation":                          []string{"UpdateEnvironment"},
		"EnvironmentName":                    []string{"env"},
		"OptionSettings.member.1.Namespace":  []string{"aws:elasticbeanstalk:application:environment"},
		"OptionSettings.member.1.OptionName": []string{"K0"},
		"OptionSettings.member.1.Value":      []string{"V0"},
		"OptionSettings.member.2.Namespace":  []string{"aws:elasticbeanstalk:application:environment"},
		"OptionSettings.member.2.OptionName": []string{"K1"},
		"OptionSettings.member.2.Value":      []string{"V1"},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		if p := r.URL.Query(); !reflect.DeepEqual(p, wantParams) {
			t.Errorf("UpdateEnvironment got params %# v, want %# v", pretty.Formatter(p), pretty.Formatter(wantParams))
		}
	})

	p := &UpdateEnvironmentParams{EnvironmentName: "env"}
	p.AddEnv("K0", "V0")
	p.AddEnv("K1", "V1")
	err := client.UpdateEnvironment(p)
	if err != nil {
		t.Errorf("UpdateEnvironment returned error: %v", err)
	}
}
