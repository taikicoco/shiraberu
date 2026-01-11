package period

import "time"

// Type は期間の種類を表す
type Type string

const (
	TypeWeek   Type = "week"
	TypeMonth  Type = "month"
	TypeCustom Type = "custom"
)

// CalcPrevious は指定された期間の直前の同等期間を計算する
func CalcPrevious(startDate, endDate time.Time, periodType Type) (time.Time, time.Time) {
	switch periodType {
	case TypeWeek:
		// Previous week (Monday to Sunday)
		prevEndDate := startDate.AddDate(0, 0, -1)
		prevStartDate := prevEndDate.AddDate(0, 0, -6)
		return prevStartDate, prevEndDate

	case TypeMonth:
		// Previous month (1st to last day)
		prevEndDate := startDate.AddDate(0, 0, -1)
		prevStartDate := time.Date(prevEndDate.Year(), prevEndDate.Month(), 1, 0, 0, 0, 0, prevEndDate.Location())
		return prevStartDate, prevEndDate

	default: // TypeCustom
		// Same duration before
		duration := endDate.Sub(startDate) + 24*time.Hour
		prevEndDate := startDate.AddDate(0, 0, -1)
		prevStartDate := prevEndDate.Add(-duration + 24*time.Hour)
		return prevStartDate, prevEndDate
	}
}
