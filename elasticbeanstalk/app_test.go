package elasticbeanstalk

import (
	"net/http"
	"testing"
)

func TestCreateApplicationVersion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
	})

	err := client.CreateApplicationVersion(&CreateApplicationVersionParams{})
	if err != nil {
		t.Errorf("CreateApplicationVersion returned error: %v", err)
	}
}
