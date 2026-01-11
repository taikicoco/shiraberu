// Package errors はアプリケーション固有のエラー型を定義する
package errors

import "errors"

// Sentinel errors for GitHub client
var (
	// ErrEmptyUsername はGitHub APIから空のユーザー名が返された場合のエラー
	ErrEmptyUsername = errors.New("empty username returned from GitHub API")

	// ErrAPIFailed はGitHub API呼び出しが失敗した場合のエラー
	ErrAPIFailed = errors.New("GitHub API call failed")
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
)
