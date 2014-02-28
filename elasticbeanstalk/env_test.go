package elasticbeanstalk

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestDescribeEnvironments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		io.WriteString(w, `
{
    "Environments": [
        {
            "ApplicationName": "app",
            "CNAME": "app-env.elasticbeanstalk.com",
            "DateCreated": "2014-02-28T00:22:21.474Z",
            "DateUpdated": "2014-02-28T00:33:47.684Z",
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
}
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

	if !reflect.DeepEqual(envs, want) {
		t.Errorf("DescribeEnvironments returned %+v, want %+v", envs, want)
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
