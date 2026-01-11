# Shiraberu リファクタリング提案

> 作成日: 2026-01-11
> 対象: 全コードベース（テストコード含む）

---

## 目次

1. [即座に修正すべき問題 (Critical)](#1-即座に修正すべき問題-critical)
2. [アーキテクチャの問題 (High)](#2-アーキテクチャの問題-high)
3. [コード品質の問題 (Medium)](#3-コード品質の問題-medium)
4. [テストの問題 (Medium)](#4-テストの問題-medium)
5. [HTML テンプレートの問題 (Medium)](#5-html-テンプレートの問題-medium)
6. [設計改善の提案 (Enhancement)](#6-設計改善の提案-enhancement)
7. [優先順位付きアクションプラン](#7-優先順位付きアクションプラン)
8. [良い点（維持すべき）](#8-良い点維持すべき)

---

## 1. 即座に修正すべき問題 (Critical)

### 1.1 `internal/pr/fetcher_test.go` - 不要な自作関数

**場所**: `internal/pr/fetcher_test.go:143-154`

```go
// 現状: 自作のcontains関数
func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

**問題**:
- `strings.Contains` で十分
- 自作関数はバグの温床になりうる
- 標準ライブラリの方がパフォーマンスも良い

**修正案**:
```go
import "strings"

// containsとcontainsHelperを削除し、strings.Containsを使用
func (m *MockPRSearcher) SearchPRs(org string, query string, dateFilter string) ([]github.PullRequest, error) {
    if m.err != nil {
        return nil, m.err
    }

    if strings.Contains(query, "is:open") {
        return m.openedPRs, nil
    }
    if strings.Contains(query, "is:merged") {
        return m.mergedPRs, nil
    }
    if strings.Contains(query, "reviewed-by:") {
        return m.reviewPRs, nil
    }

    return nil, nil
}
```

---

### 1.2 タイムゾーン定義の重複

**場所**: 複数箇所で同一定義が重複

| ファイル | 行 |
|----------|-----|
| `internal/pr/fetcher.go` | 38 |
| `internal/pr/fetcher.go` | 75 |
| `internal/demo/demo.go` | 56 |
| `internal/pr/fetcher_test.go` | 11, 165 |
| `internal/render/html_test.go` | 70, 145, 217, 382, 398, 478, 524, 546, 604 |
| `internal/server/server_test.go` | 19, 77, 120, 157, 222 |

```go
// 現状: 各所で同じ定義
jst := time.FixedZone("JST", 9*60*60)
```

**問題**:
- DRY原則に違反
- タイムゾーン変更時に複数箇所の修正が必要
- タイポのリスク

**修正案**: 共通パッケージに定数として定義

```go
// internal/timezone/timezone.go
package timezone

import "time"

// JST は日本標準時 (UTC+9) のタイムゾーン
var JST = time.FixedZone("JST", 9*60*60)
```

使用側:
```go
import "github.com/taikicoco/shiraberu/internal/timezone"

date := time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST)
```

---

### 1.3 エラーの無視

**場所**: `internal/prompt/prompt.go`

```go
// line 235
input, _ := reader.ReadString('\n')

// line 269
input, _ := reader.ReadString('\n')
```

**問題**:
- EOF やその他のエラーを検知できない
- パイプ入力時に予期しない動作の可能性

**修正案**:
```go
input, err := reader.ReadString('\n')
if err != nil && err != io.EOF {
    return time.Time{}, time.Time{}, true // エラー時はバックに戻る
}
```

---

## 2. アーキテクチャの問題 (High)

### 2.1 `main.go` の責務過多

**現状**: `calcPreviousPeriod` 関数が `main.go` に存在 (lines 112-133)

```go
func calcPreviousPeriod(startDate, endDate time.Time, periodType prompt.PeriodType) (time.Time, time.Time) {
    switch periodType {
    case prompt.PeriodTypeWeek:
        // ...
    case prompt.PeriodTypeMonth:
        // ...
    default:
        // ...
    }
}
```

**問題**:
- ドメインロジックがエントリーポイントに混在
- テストが `main_test.go` に配置される不自然さ
- 再利用が困難

**修正案**: 新パッケージへ抽出

```
internal/period/
├── period.go       # PeriodType, CalcPreviousPeriod
└── period_test.go  # テスト移動
```

```go
// internal/period/period.go
package period

import "time"

type Type string

const (
    TypeWeek   Type = "week"
    TypeMonth  Type = "month"
    TypeCustom Type = "custom"
)

// CalcPrevious は指定された期間の直前の同等期間を計算する
func CalcPrevious(startDate, endDate time.Time, periodType Type) (time.Time, time.Time) {
    // 実装
}
```

---

### 2.2 インターフェースの配置

**現状**: `PRSearcher` インターフェースが `internal/github/client.go` に定義 (lines 14-18)

```go
// PRSearcher is an interface for searching PRs (for mocking in tests)
type PRSearcher interface {
    Username() string
    SearchPRs(org string, query string, dateFilter string) ([]PullRequest, error)
}
```

**問題**:
- Goの慣習では、インターフェースは「利用する側」で定義すべき
- "Accept interfaces, return structs" の原則
- 現状では github パッケージが自身のインターフェースを定義している

**修正案**: `internal/pr/fetcher.go` に移動

```go
// internal/pr/fetcher.go
package pr

import "github.com/taikicoco/shiraberu/internal/github"

// PRSearcher はPR検索機能を抽象化するインターフェース
type PRSearcher interface {
    Username() string
    SearchPRs(org string, query string, dateFilter string) ([]github.PullRequest, error)
}

type Fetcher struct {
    client PRSearcher
}
```

`internal/github/client.go` からはインターフェース定義と `var _ PRSearcher = (*Client)(nil)` を削除。

---

### 2.3 `internal/render/html.go` の肥大化

**現状**: 396行の単一ファイル

**責務の混在**:
- 型定義 (Summary, DailyStat, WeeklyStat, MonthlyStat, RepoStat, DayJSON, PRJSON, HTMLData)
- 統計計算 (calcSummary, calcDailyStats, calcWeeklyStats, calcMonthlyStats, calcRepoStats, calcSummaryDiff)
- JSON変換 (convertToDaysJSON, convertPRsToJSON)
- レンダリング (RenderHTML)

**修正案**: 責務ごとにファイル分割

```
internal/render/
├── html.go         # RenderHTML のみ (~50行)
├── markdown.go     # RenderMarkdown のみ (現状維持)
├── stats.go        # calc* 関数群 (~200行)
├── convert.go      # convertTo* 関数群 (~40行)
└── types.go        # 型定義 (~100行)
```

---

## 3. コード品質の問題 (Medium)

### 3.1 `prompt.PeriodType` の場所

**現状**: `internal/prompt/prompt.go` に定義 (lines 17-23)

```go
type PeriodType string

const (
    PeriodTypeWeek   PeriodType = "week"
    PeriodTypeMonth  PeriodType = "month"
    PeriodTypeCustom PeriodType = "custom"
)
```

**問題**:
- `main.go` の `calcPreviousPeriod` で使用されている
- prompt はUI層だが、PeriodType はドメイン概念
- 依存関係が不自然

**修正案**: `internal/period/period.go` に移動（2.1と併せて実施）

---

### 3.2 マジックナンバー

**場所**: 複数箇所

| ファイル | 行 | 値 | 意味 |
|----------|-----|-----|------|
| `internal/spinner/spinner.go` | 41 | `80 * time.Millisecond` | スピナーの更新間隔 |
| `internal/server/server.go` | 48 | `500 * time.Millisecond` | ブラウザ起動の待機時間 |
| `internal/server/server.go` | 17 | `"7777"` | デフォルトポート（定数化済みだが） |

**修正案**: 定数化

```go
// internal/spinner/spinner.go
const spinnerUpdateInterval = 80 * time.Millisecond

// internal/server/server.go
const browserOpenDelay = 500 * time.Millisecond
```

---

### 3.3 GraphQL クエリの管理

**現状**: `internal/github/client.go` に文字列リテラルとして埋め込み (lines 52-81)

```go
const searchQuery = `
query($q: String!, $cursor: String) {
  search(query: $q, type: ISSUE, first: 100, after: $cursor) {
    ...
  }
}
`
```

**問題**:
- IDEのシンタックスハイライトが効かない
- 複数クエリが増えた場合の管理が困難

**修正案**（将来的）:
```
internal/github/
├── client.go
├── types.go
└── queries/
    └── search.graphql
```

embedで読み込み:
```go
//go:embed queries/*.graphql
var queryFS embed.FS
```

---

## 4. テストの問題 (Medium)

### 4.1 `time.Sleep` 依存のテスト

**場所**: `internal/spinner/spinner_test.go`

```go
// lines 19, 39, 59, 78, 93
time.Sleep(100 * time.Millisecond)
time.Sleep(50 * time.Millisecond)
```

**問題**:
- CI環境でフレーキー（不安定）になる可能性
- テスト実行時間の増加
- 実際の動作を保証しない

**修正案**: テスタブルな設計に変更

```go
// Spinner にテスト用フックを追加
type Spinner struct {
    // ...
    ticker *time.Ticker  // 外部から注入可能に
}

// テスト時
func TestSpinner_StartStop(t *testing.T) {
    s := &Spinner{
        message: "Loading...",
        writer:  &buf,
        // ...
    }

    // チャネルで同期を取る
    started := make(chan struct{})
    s.onStart = func() { close(started) }

    s.Start()
    <-started  // 開始を待つ
    s.Stop()
}
```

---

### 4.2 テストのモック方法

**現状**: `execCommand` グローバル変数を差し替え

```go
// internal/github/client.go:12
var execCommand = exec.Command

// internal/github/client_test.go:111
execCommand = fakeExecCommand
```

**問題**:
- テストの並列実行で競合リスク (`t.Parallel()` 使用時)
- グローバル状態の変更は予期しない副作用を生む

**修正案**: 依存性注入パターン

```go
// internal/github/client.go
type CommandExecutor interface {
    Execute(name string, args ...string) ([]byte, error)
}

type execCommandExecutor struct{}

func (e *execCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
    return exec.Command(name, args...).Output()
}

type Client struct {
    username string
    executor CommandExecutor
}

func NewClient() (*Client, error) {
    return NewClientWithExecutor(&execCommandExecutor{})
}

func NewClientWithExecutor(executor CommandExecutor) (*Client, error) {
    // ...
}
```

---

### 4.3 テストカバレッジの欠落

**テストされていない関数**:

| ファイル | 関数 | 理由 |
|----------|------|------|
| `internal/prompt/prompt.go` | `Run` | インタラクティブ |
| `internal/prompt/prompt.go` | `promptText` | stdin依存 |
| `internal/prompt/prompt.go` | `confirmDateRange` | 再帰的なインタラクション |
| `internal/prompt/prompt.go` | `promptSelect` | promptui依存 |

**修正案**: IO を抽象化

```go
type PromptIO interface {
    ReadLine() (string, error)
    Select(label string, options []string) (int, error)
}

func RunWithIO(cfg *config.Config, io PromptIO) (*Options, error) {
    // ...
}
```

---

## 5. HTML テンプレートの問題 (Medium)

### 5.1 巨大な単一ファイル

**現状**: `internal/render/templates/report.html` = 1117行

| セクション | 行数（概算） |
|-----------|-------------|
| HTML構造 | ~100行 |
| CSS | ~390行 |
| JavaScript | ~500行 |
| Goテンプレート | ~130行 |

**問題**:
- 保守性の低下
- コードレビューが困難
- CSS/JSの単体テスト不可

**修正案（開発体験向上版）**:

```
internal/render/
├── templates/
│   ├── report.html      # 最小限のHTML構造
│   └── assets/
│       ├── styles.css   # 開発時は分離
│       └── scripts.js   # 開発時は分離
└── embed.go             # ビルド時にインライン化
```

ビルドスクリプト:
```makefile
.PHONY: embed-assets
embed-assets:
	# CSSとJSをHTMLにインライン化
	go run ./tools/embed-assets
```

---

### 5.2 JavaScript のエラーハンドリング

**場所**: `report.html` JavaScript部分

```javascript
// line 996-999
function addDays(dateStr, days) {
    const d = new Date(dateStr);
    d.setDate(d.getDate() + days);
    return d.toISOString().split('T')[0];
}
```

**問題**:
- 無効な日付文字列で `Invalid Date` になる
- `toISOString()` が例外を投げる可能性

**修正案**:
```javascript
function addDays(dateStr, days) {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) {
        console.error('Invalid date:', dateStr);
        return dateStr; // フォールバック
    }
    d.setDate(d.getDate() + days);
    return d.toISOString().split('T')[0];
}
```

---

## 6. 設計改善の提案 (Enhancement)

### 6.1 パッケージ構造の再編

**現状**:
```
internal/
├── config/
├── demo/
├── github/
├── pr/
├── prompt/
├── render/
├── server/
└── spinner/
```

**提案**:
```
internal/
├── config/         # 設定 (現状維持)
├── github/         # GitHub API クライアント
│   ├── client.go
│   └── types.go
├── period/         # 期間計算 (新規)
│   ├── period.go   # Type, CalcPrevious
│   └── period_test.go
├── timezone/       # タイムゾーン (新規)
│   └── timezone.go # var JST
├── pr/             # PR ドメイン
│   ├── types.go    # DailyPRs, Report
│   ├── fetcher.go  # PRSearcher interface, Fetcher
│   └── fetcher_test.go
├── render/         # レンダリング
│   ├── types.go    # Summary, DailyStat など
│   ├── stats.go    # 統計計算
│   ├── html.go     # HTML レンダリング
│   ├── markdown.go # Markdown レンダリング
│   └── templates/
├── server/         # HTTP サーバー (現状維持)
├── spinner/        # スピナー (現状維持)
├── prompt/         # プロンプト (PeriodType削除)
└── demo/           # デモデータ (現状維持)
```

---

### 6.2 CLI フレームワークの導入検討

**現状**: 標準 `flag` パッケージ

```go
var demoMode = flag.Bool("demo", false, "Run with demo data")
```

**提案**: `cobra` または `urfave/cli` の導入

**メリット**:
- サブコマンドの追加が容易 (`shiraberu report`, `shiraberu config`)
- 自動ヘルプ生成
- シェル補完のサポート
- 設定ファイル（Viper）との統合

**例**:
```go
var rootCmd = &cobra.Command{
    Use:   "shiraberu",
    Short: "GitHub PR activity visualizer",
}

var reportCmd = &cobra.Command{
    Use:   "report",
    Short: "Generate PR report",
    RunE:  runReport,
}

func init() {
    rootCmd.AddCommand(reportCmd)
    reportCmd.Flags().Bool("demo", false, "Run with demo data")
}
```

---

### 6.3 ロギングの導入

**現状**: `fmt.Fprintf(os.Stderr, ...)`

```go
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
```

**提案**: `log/slog` (Go 1.21+) の採用

```go
import "log/slog"

// 初期化
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// 使用
slog.Error("failed to fetch PRs", "org", org, "error", err)
slog.Info("report generated", "path", outputPath)
```

**メリット**:
- 構造化ログ
- ログレベルの制御
- JSON出力への切り替えが容易

---

### 6.4 設定の拡張

**現状**: 環境変数のみ

```go
cfg := &Config{
    Org:       os.Getenv("SHIRABERU_ORG"),
    Format:    getEnvOrDefault("SHIRABERU_FORMAT", "markdown"),
    OutputDir: getEnvOrDefault("SHIRABERU_OUTPUT_DIR", "./output"),
}
```

**提案**: 設定ファイルのサポート追加

```yaml
# ~/.config/shiraberu/config.yaml
org: my-org
format: html
output_dir: ~/reports
timezone: Asia/Tokyo
```

**実装**:
```go
import "github.com/spf13/viper"

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.AddConfigPath("$HOME/.config/shiraberu")
    viper.AutomaticEnv()
    viper.SetEnvPrefix("SHIRABERU")

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    return &Config{
        Org:       viper.GetString("org"),
        Format:    viper.GetString("format"),
        OutputDir: viper.GetString("output_dir"),
    }, nil
}
```

---

## 7. 優先順位付きアクションプラン

| 優先度 | 項目 | 工数 | 影響範囲 |
|--------|------|------|----------|
| **P0** | `strings.Contains` への置換 | 5分 | fetcher_test.go |
| **P0** | タイムゾーン定数の共通化 | 15分 | 全体 |
| **P0** | エラー無視の修正 | 10分 | prompt.go |
| **P1** | `calcPreviousPeriod` を別パッケージへ | 30分 | main.go, main_test.go |
| **P1** | `PeriodType` の移動 | 20分 | prompt.go, main.go |
| **P1** | `PRSearcher` インターフェースの移動 | 15分 | github/client.go, pr/fetcher.go |
| **P2** | render パッケージの分割 | 1-2時間 | render/ |
| **P2** | テストの `time.Sleep` 排除 | 1時間 | spinner_test.go |
| **P2** | マジックナンバーの定数化 | 20分 | 複数ファイル |
| **P3** | execCommand のDI化 | 1時間 | github/client.go, client_test.go |
| **P3** | prompt のテスタビリティ向上 | 2時間 | prompt/ |
| **P3** | HTML テンプレートの分離 | 2-3時間 | render/templates/ |
| **P4** | CLI フレームワーク導入 | 2-3時間 | main.go |
| **P4** | slog 導入 | 1時間 | 全体 |
| **P4** | 設定ファイルサポート | 1-2時間 | config/ |

---

## 8. 良い点（維持すべき）

### 8.1 テストカバレッジが高い

ほぼすべてのパッケージにテストファイルが存在:
- `config/config_test.go`
- `github/client_test.go`
- `pr/fetcher_test.go`
- `prompt/prompt_test.go`
- `render/html_test.go`
- `server/server_test.go`
- `spinner/spinner_test.go`
- `demo/demo_test.go`
- `main_test.go`

### 8.2 インターフェースによる抽象化

`PRSearcher` インターフェースにより、テスト時のモックが容易:
```go
type MockPRSearcher struct {
    username  string
    openedPRs []github.PullRequest
    // ...
}
```

### 8.3 embed による静的ファイル管理

HTMLテンプレートが単一バイナリに含まれる:
```go
//go:embed templates/*.html
var templateFS embed.FS
```

### 8.4 テーブル駆動テスト

一貫したテストスタイル:
```go
tests := []struct {
    name string
    input string
    want string
}{
    {"case1", "input1", "output1"},
    // ...
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // ...
    })
}
```

### 8.5 エラーラップの適切な使用

```go
return nil, fmt.Errorf("failed to get GitHub username: %w", err)
```

### 8.6 デモモードの存在

API呼び出しなしでUIを確認できる:
```bash
go run . --demo
```

### 8.7 golangci-lint の設定

`.golangci.yml` で適切なリンター設定:
```yaml
linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - ineffassign
```

---

## 付録: 参考リンク

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
