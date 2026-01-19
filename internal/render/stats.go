package render

import (
	"fmt"
	"sort"
	"time"

	"github.com/taikicoco/shiraberu/internal/pr"
)

func calcSummary(report *pr.Report) Summary {
	var s Summary
	for _, day := range report.Days {
		s.OpenedCount += len(day.Opened)
		s.DraftCount += len(day.Draft)
		s.MergedCount += len(day.Merged)
		s.ReviewedCount += len(day.Reviewed)
		// Additions/Deletions are only counted for merged PRs
		for _, p := range day.Merged {
			s.Additions += p.Additions
			s.Deletions += p.Deletions
		}
	}
	return s
}

func calcDailyStats(report *pr.Report) []DailyStat {
	// 期間内の全日を含むマップを作成
	dayMap := make(map[string]*DailyStat)

	// まず期間内の全日を0で初期化
	for d := report.StartDate; !d.After(report.EndDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayMap[dateStr] = &DailyStat{
			Date: dateStr,
		}
	}

	// PRがある日のデータを埋める
	for _, day := range report.Days {
		dateStr := day.Date.Format("2006-01-02")
		stat, ok := dayMap[dateStr]
		if !ok {
			stat = &DailyStat{Date: dateStr}
			dayMap[dateStr] = stat
		}

		var additions, deletions int
		for _, p := range day.Opened {
			additions += p.Additions
			deletions += p.Deletions
		}
		for _, p := range day.Draft {
			additions += p.Additions
			deletions += p.Deletions
		}
		for _, p := range day.Merged {
			additions += p.Additions
			deletions += p.Deletions
		}

		stat.OpenedCount = len(day.Opened)
		stat.DraftCount = len(day.Draft)
		stat.MergedCount = len(day.Merged)
		stat.ReviewedCount = len(day.Reviewed)
		stat.Additions = additions
		stat.Deletions = deletions
		stat.TotalPRs = len(day.Opened) + len(day.Draft) + len(day.Merged) + len(day.Reviewed)
	}

	// スライスに変換
	stats := make([]DailyStat, 0, len(dayMap))
	for _, stat := range dayMap {
		stats = append(stats, *stat)
	}

	// グラフ用に日付昇順（左=過去、右=最新）にソート
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Date < stats[j].Date
	})

	return stats
}

func calcWeeklyStats(report *pr.Report) []WeeklyStat {
	weekMap := make(map[string]*WeeklyStat)

	// 期間内の全週を0で初期化
	for d := report.StartDate; !d.After(report.EndDate); d = d.AddDate(0, 0, 1) {
		year, week := d.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		if _, ok := weekMap[weekKey]; !ok {
			// Calculate week start date (Monday) for label
			weekStart := d.AddDate(0, 0, -int(d.Weekday())+1)
			if d.Weekday() == 0 { // Sunday
				weekStart = d.AddDate(0, 0, -6)
			}
			weekEnd := weekStart.AddDate(0, 0, 6) // Sunday
			weekLabel := fmt.Sprintf("%d/%d 〜 %d/%d", weekStart.Month(), weekStart.Day(), weekEnd.Month(), weekEnd.Day())
			weekMap[weekKey] = &WeeklyStat{
				Week:      weekLabel,
				StartDate: weekStart.Format("2006-01-02"),
				EndDate:   weekEnd.Format("2006-01-02"),
			}
		}
	}

	// PRがある日のデータを埋める
	for _, day := range report.Days {
		year, week := day.Date.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		if stat, ok := weekMap[weekKey]; ok {
			stat.OpenedCount += len(day.Opened)
			stat.DraftCount += len(day.Draft)
			stat.MergedCount += len(day.Merged)
			stat.ReviewedCount += len(day.Reviewed)
		}
	}

	// スライスに変換してソート
	stats := make([]WeeklyStat, 0, len(weekMap))
	keys := make([]string, 0, len(weekMap))
	for k := range weekMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		stats = append(stats, *weekMap[k])
	}

	return stats
}

func calcMonthlyStats(report *pr.Report) []MonthlyStat {
	monthMap := make(map[string]*MonthlyStat)

	// 期間内の全月を0で初期化
	for d := report.StartDate; !d.After(report.EndDate); {
		monthKey := d.Format("2006-01")
		monthLabel := d.Format("Jan 2006")

		if _, ok := monthMap[monthKey]; !ok {
			// Calculate month start and end dates
			monthStart := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
			monthEnd := monthStart.AddDate(0, 1, -1) // Last day of month
			monthMap[monthKey] = &MonthlyStat{
				Month:     monthLabel,
				StartDate: monthStart.Format("2006-01-02"),
				EndDate:   monthEnd.Format("2006-01-02"),
			}
		}

		// 次の月の1日へ移動
		d = time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, 0, d.Location())
	}

	// PRがある日のデータを埋める
	for _, day := range report.Days {
		monthKey := day.Date.Format("2006-01")

		if stat, ok := monthMap[monthKey]; ok {
			stat.OpenedCount += len(day.Opened)
			stat.DraftCount += len(day.Draft)
			stat.MergedCount += len(day.Merged)
			stat.ReviewedCount += len(day.Reviewed)
		}
	}

	// スライスに変換してソート
	stats := make([]MonthlyStat, 0, len(monthMap))
	keys := make([]string, 0, len(monthMap))
	for k := range monthMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		stats = append(stats, *monthMap[k])
	}

	return stats
}

func calcRepoStats(report *pr.Report) []RepoStat {
	repoCount := make(map[string]int)
	for _, day := range report.Days {
		for _, p := range day.Opened {
			repoCount[p.Repository]++
		}
		for _, p := range day.Draft {
			repoCount[p.Repository]++
		}
		for _, p := range day.Merged {
			repoCount[p.Repository]++
		}
		for _, p := range day.Reviewed {
			repoCount[p.Repository]++
		}
	}

	stats := make([]RepoStat, 0, len(repoCount))
	for repo, count := range repoCount {
		stats = append(stats, RepoStat{
			Repository: repo,
			Count:      count,
		})
	}

	// 件数の多い順にソート
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

func calcSummaryDiff(current Summary, previousReport *pr.Report) SummaryDiff {
	if previousReport == nil {
		return SummaryDiff{HasPrevious: false}
	}

	prev := calcSummary(previousReport)
	return SummaryDiff{
		OpenedDiff:   current.OpenedCount - prev.OpenedCount,
		DraftDiff:    current.DraftCount - prev.DraftCount,
		MergedDiff:   current.MergedCount - prev.MergedCount,
		ReviewedDiff: current.ReviewedCount - prev.ReviewedCount,
		HasPrevious:  true,
	}
}
