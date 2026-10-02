package iamservicelogic

import "testing"

func TestPermissionMatches(t *testing.T) {
	for _, test := range []struct {
		granted  string
		required string
		want     bool
	}{
		{"user.read", "user.read", true},
		{"user.*", "user.update", true},
		{"*.read", "config.read", true},
		{"*", "attachment.delete", true},
		{"user.read", "user.delete", false},
	} {
		if got := permissionMatches(test.granted, test.required); got != test.want {
			t.Errorf("permissionMatches(%q, %q) = %v, want %v", test.granted, test.required, got, test.want)
		}
	}
}
