package period

import (
	"testing"
	"time"
)

func TestCalcPrevious_Week(t *testing.T) {
	tests := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		wantPrevStart time.Time
		wantPrevEnd   time.Time
	}{
		{
			name:          "this week Mon-Sun",
			startDate:     time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),  // Monday
			endDate:       time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC), // Sunday
			wantPrevStart: time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "week crossing year",
			startDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := CalcPrevious(tt.startDate, tt.endDate, TypeWeek)
			if !gotStart.Equal(tt.wantPrevStart) {
				t.Errorf("prevStartDate: got %v, want %v", gotStart, tt.wantPrevStart)
			}
			if !gotEnd.Equal(tt.wantPrevEnd) {
				t.Errorf("prevEndDate: got %v, want %v", gotEnd, tt.wantPrevEnd)
			}
		})
	}
}

func TestCalcPrevious_Month(t *testing.T) {
	tests := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		wantPrevStart time.Time
		wantPrevEnd   time.Time
	}{
		{
			name:          "december to november",
			startDate:     time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "november to october",
			startDate:     time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "january to december (year crossing)",
			startDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "march to february (leap year)",
			startDate:     time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := CalcPrevious(tt.startDate, tt.endDate, TypeMonth)
			if !gotStart.Equal(tt.wantPrevStart) {
				t.Errorf("prevStartDate: got %v, want %v", gotStart, tt.wantPrevStart)
			}
			if !gotEnd.Equal(tt.wantPrevEnd) {
				t.Errorf("prevEndDate: got %v, want %v", gotEnd, tt.wantPrevEnd)
			}
		})
	}
}

func TestCalcPrevious_Custom(t *testing.T) {
	tests := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		wantPrevStart time.Time
		wantPrevEnd   time.Time
	}{
		{
			name:          "10 days custom range",
			startDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 24, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "single day",
			startDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			wantPrevStart: time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC),
			wantPrevEnd:   time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := CalcPrevious(tt.startDate, tt.endDate, TypeCustom)
			if !gotStart.Equal(tt.wantPrevStart) {
				t.Errorf("prevStartDate: got %v, want %v", gotStart, tt.wantPrevStart)
			}
			if !gotEnd.Equal(tt.wantPrevEnd) {
				t.Errorf("prevEndDate: got %v, want %v", gotEnd, tt.wantPrevEnd)
			}
		})
	}
}
