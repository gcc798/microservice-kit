package database

import "testing"

func TestCloseNil(t *testing.T) {
	if err := Close(nil); err != nil {
		t.Fatal(err)
	}
}
