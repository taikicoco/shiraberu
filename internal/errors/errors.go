// Package errors はアプリケーション固有のエラー型を定義する
package errors

import "errors"

// Sentinel errors for GitHub client
var (
	// ErrEmptyUsername はGitHub APIから空のユーザー名が返された場合のエラー
	ErrEmptyUsername = errors.New("empty username returned from GitHub API")

	// ErrAPIFailed はGitHub API呼び出しが失敗した場合のエラー
	ErrAPIFailed = errors.New("GitHub API call failed")

	// ErrTooManyResults は検索結果がGitHubの上限（1,000件）を超えた場合のエラー
	ErrTooManyResults = errors.New("search results exceed the GitHub limit of 1000")
)

// Sentinel errors for config
var (
	// ErrConfigNotFound は設定ファイルが見つからない場合のエラー
	ErrConfigNotFound = errors.New("config file not found")

	// ErrInvalidConfig は設定ファイルの形式が無効な場合のエラー
	ErrInvalidConfig = errors.New("invalid config format")
)

// Sentinel errors for prompt
var (
	// ErrInvalidDate は日付形式が無効な場合のエラー
	ErrInvalidDate = errors.New("invalid date format")

	// ErrOrgRequired は組織名が必須だが空の場合のエラー
	ErrOrgRequired = errors.New("organization is required")

	// ErrSinceRequired は対話なしの実行で -since が指定されていない場合のエラー
	ErrSinceRequired = errors.New("-since is required when running without prompts")

	// ErrInvalidDateRange は開始日が終了日より後の場合のエラー
	ErrInvalidDateRange = errors.New("-since must not be after -until")

	// ErrUnsupportedFormat は対話なしの実行で扱えない出力形式の場合のエラー
	ErrUnsupportedFormat = errors.New("unsupported format")
)
