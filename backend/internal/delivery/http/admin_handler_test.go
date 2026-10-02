package http

import (
	"net/http/httptest"
	"testing"
)

func TestParseAttendanceFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=25&offset=5&status=late&date_from=2026-10-01&date_to=2026-10-02", nil)
	f, err := parseAttendanceFilter(r)
	if err != nil {
		t.Fatal(err)
	}
	if f.Limit != 25 || f.Offset != 5 || f.Status != "late" || f.DateFrom == nil || f.DateTo == nil {
		t.Fatalf("unexpected filter: %+v", f)
	}
}

func TestParseAttendanceFilterRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"status=unknown", "date_from=01-10-2026", "location_id=nope"} {
		r := httptest.NewRequest("GET", "/?"+raw, nil)
		if _, err := parseAttendanceFilter(r); err == nil {
			t.Fatalf("expected error for %s", raw)
		}
	}
}
