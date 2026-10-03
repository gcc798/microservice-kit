package worker

import (
	"testing"
	"time"
)

func TestScheduleRejectsOverlappingClaim(t *testing.T) {
	if _, err := Schedule("0 * * * * *", time.Minute); err == nil {
		t.Fatal("overlapping claim window was accepted")
	}
	if _, err := Schedule("0 * * * * *", 30*time.Second); err != nil {
		t.Fatal(err)
	}
}
