package server

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

type Xdata struct {
	Expire int64
}

func TestSubmarshalling(t *testing.T) {
	jsonData := "{\"expire\":404}"
	var sub Xdata
	dec := json.NewDecoder(strings.NewReader(jsonData))
	for {
		err := dec.Decode(&sub)
		if err == io.EOF {
			break
		} else if err != nil {
			sub.Expire = -1
			t.Errorf("Error in Json Data")
		}
	}

	if sub.Expire != 404 {
		t.Errorf("Could not Marshall Json Data to Go object")
	}
}
