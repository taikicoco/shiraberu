package timezone

import (
	"testing"
	"time"
)

func TestJST(t *testing.T) {
	if JST == nil {
		t.Fatal("JST should not be nil")
	}

	// JST は UTC+9 (32400秒)
	// Use Zone method with a specific time to get name and offset
	testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, JST)
	name, offset := testTime.Zone()
	if name != "JST" {
		t.Errorf("JST name: got %q, want %q", name, "JST")
	}
	if offset != 9*60*60 {
		t.Errorf("JST offset: got %d, want %d", offset, 9*60*60)
	}
}

func TestJST_TimeConversion(t *testing.T) {
	// UTC 2025-01-01 00:00:00 は JST 2025-01-01 09:00:00
	utc := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	jst := utc.In(JST)

	if jst.Hour() != 9 {
		t.Errorf("JST hour: got %d, want 9", jst.Hour())
	}
	if jst.Day() != 1 {
		t.Errorf("JST day: got %d, want 1", jst.Day())
	}
}

func TestJST_DateBoundary(t *testing.T) {
	// UTC 2025-01-01 15:00:00 は JST 2025-01-02 00:00:00 (日付が変わる)
	utc := time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)
	jst := utc.In(JST)

	if jst.Day() != 2 {
		t.Errorf("JST day: got %d, want 2", jst.Day())
	}
	if jst.Hour() != 0 {
		t.Errorf("JST hour: got %d, want 0", jst.Hour())
	}
}
