package common

import "testing"

func TestRegexpMatched(t *testing.T) {
	matched, err := RegexpMatched(`^[a-z]+[a-z\-]*[a-z]$`, "acs-dcs-app")
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("matched:",matched)
	}
}
