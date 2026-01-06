# shiraberu

GitHub PR活動をログとして可視化するCLIツール

## Features

- 自分がオープン/マージ/レビューしたPRを日別に表示
- Draft PRを分離して表示
- コード変更量（追加/削除行数）、コメント数を表示
- HTML/Markdown/ブラウザ表示に対応
- GitHub GraphQL APIで高速なデータ取得

## Requirements

- Go 1.21+
- [GitHub CLI](https://cli.github.com/) (`gh`) - 認証済みであること

## Installation

```bash
go install github.com/taikicoco/shiraberu/cmd/shiraberu@latest
```

## Usage

```bash
shiraberu
```

## Development

```bash
git clone https://github.com/taikicoco/shiraberu.git
cd shiraberu
make run
```

対話形式で以下を選択:
1. Organization
2. 期間（今日/昨日/今週/先週/カスタム）
3. 出力形式（Markdown/HTML/ブラウザ）
