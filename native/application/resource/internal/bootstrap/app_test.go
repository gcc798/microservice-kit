package bootstrap

import "testing"

func TestAppCloseZeroValueIsSafe(t *testing.T) {
	var app App
	if err := app.Close(); err != nil {
		t.Fatal(err)
	}
	if err := app.Close(); err != nil {
		t.Fatal(err)
	}
}
