package render

import "github.com/taikicoco/shiraberu/internal/pr"

// Summary はPRの集計データ
type Summary struct {
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
	Additions     int
	Deletions     int
}

// SummaryDiff は前期間との差分
type SummaryDiff struct {
	OpenedDiff   int
	DraftDiff    int
	MergedDiff   int
	ReviewedDiff int
	HasPrevious  bool // 前期間データがあるかどうか
}

// DailyStat はグラフ用の日別統計データ
type DailyStat struct {
	Date          string // "2006-01-02" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
	Additions     int
	Deletions     int
	TotalPRs      int // 日別詳細のサマリー表示用
}

// WeeklyStat は週別統計データ
type WeeklyStat struct {
	Week          string // "1/1 〜 1/7" 形式
	StartDate     string // "2006-01-02" 形式
	EndDate       string // "2006-01-02" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
}

// MonthlyStat は月別統計データ
type MonthlyStat struct {
	Month         string // "Jan 2006" 形式
	StartDate     string // "2006-01-02" 形式
	EndDate       string // "2006-01-02" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
}

// RepoStat はリポジトリ別の統計データ
type RepoStat struct {
	Repository string
	Count      int
}

// DayJSON はJavaScript用の日別データ
type DayJSON struct {
	Date     string   `json:"date"`
	Opened   []PRJSON `json:"opened"`
	Draft    []PRJSON `json:"draft"`
	Merged   []PRJSON `json:"merged"`
	Reviewed []PRJSON `json:"reviewed"`
}

// PRJSON はJavaScript用のPRデータ
type PRJSON struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Repository string `json:"repository"`
	State      string `json:"state"`
	IsDraft    bool   `json:"isDraft"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	Comments   int    `json:"comments"`
}

// HTMLData はHTMLテンプレート用のデータ
type HTMLData struct {
	Report            *pr.Report
	Summary           Summary
	SummaryDiff       SummaryDiff
	DailyStats        []DailyStat
	WeeklyStats       []WeeklyStat
	MonthlyStats      []MonthlyStat
	RepoStats         []RepoStat
	Weekdays          []string
	PeriodLabel       string
	DaysJSON          []DayJSON
	OriginalStartDate string
	OriginalEndDate   string
}
