package tests

import (
	"testing"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

func TestParsePeriod(t *testing.T) {
	cases := map[string]valueobject.Period{
		"":      valueobject.PeriodDay,
		"day":   valueobject.PeriodDay,
		"WEEK":  valueobject.PeriodWeek,
		"month": valueobject.PeriodMonth,
	}
	for in, want := range cases {
		got, err := valueobject.ParsePeriod(in)
		if err != nil {
			t.Fatalf("ParsePeriod(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("ParsePeriod(%q)=%q want %q", in, got, want)
		}
	}
	if _, err := valueobject.ParsePeriod("tick"); err == nil {
		t.Fatal("expected error for unsupported period")
	}
}

func TestParseAdjustmentDefaultsPre(t *testing.T) {
	got, err := valueobject.ParseAdjustment("")
	if err != nil {
		t.Fatalf("ParseAdjustment empty: %v", err)
	}
	if got != valueobject.AdjustmentPre {
		t.Fatalf("expected AdjustmentPre, got %q", got)
	}
}

func TestParseDate(t *testing.T) {
	if _, err := valueobject.ParseDate("2026-01-15"); err != nil {
		t.Fatalf("ParseDate dash: %v", err)
	}
	if _, err := valueobject.ParseDate("20260115"); err != nil {
		t.Fatalf("ParseDate compact: %v", err)
	}
	if _, err := valueobject.ParseDate("not-a-date"); err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestDateRangeValidate(t *testing.T) {
	r := valueobject.DateRange{}
	if err := r.Validate(); err != nil {
		t.Fatalf("zero range should be valid: %v", err)
	}
	startOnly, _ := valueobject.ParseDate("2026-01-01")
	if err := (valueobject.DateRange{Start: startOnly}).Validate(); err != nil {
		t.Fatalf("start-only range should be valid: %v", err)
	}
	endOnly, _ := valueobject.ParseDate("2026-01-31")
	if err := (valueobject.DateRange{End: endOnly}).Validate(); err != nil {
		t.Fatalf("end-only range should be valid: %v", err)
	}
	start, _ := valueobject.ParseDate("2026-01-01")
	end, _ := valueobject.ParseDate("2025-12-31")
	bad := valueobject.DateRange{Start: start, End: end}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected error for end before start")
	}
}
