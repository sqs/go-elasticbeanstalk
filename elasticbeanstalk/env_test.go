package elasticbeanstalk

import (
	"net/http"
	"testing"
)

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
