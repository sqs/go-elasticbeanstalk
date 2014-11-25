package elasticbeanstalk

import (
	"encoding/json"
	"testing"
)

func TestTime_UnmarshalJSON(t *testing.T) {
	jsonStr := "1.415215656E9"
	var tm Time
	if err := json.Unmarshal([]byte(jsonStr), &tm); err != nil {
		t.Error(err)
	}

	want := mustParseTime(t, "2014-11-05T19:27:36Z").UTC()
	if !tm.Equal(want) {
		t.Errorf("got %v, want %v", tm, want)
	}
}

func TestTime_MarshalJSON(t *testing.T) {
	tm := mustParseTime(t, "2014-11-05T19:27:36Z")
	jsonStr, err := json.Marshal(tm)
	if err != nil {
		t.Error(err)
	}

	want := "1.415215656E9"
	if string(jsonStr) != want {
		t.Errorf("got %s, want %v", jsonStr, want)
	}
}
