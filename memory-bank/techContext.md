# Tech Context - 使用技術の詳細

## プログラミング言語
### Go (1.21+)
- **選定理由**:
  - Linuxシステムプログラミングに適している
  - 静的型付けによる安全性
  - 優れた並行処理サポート
  - netlinkライブラリの充実
- **必要な言語機能**:
  - Generics (Go 1.18+) - Result[T,E]型の実装 ✅
  - Error wrapping - エラーコンテキストの保持 ✅
  - Context - キャンセレーションとタイムアウト ✅

## 確定済み主要ライブラリ

### 1. Netlink通信
**vishvananda/netlink v1.2.1** ✅
- Linux kernelとの通信実装済み
- TC操作の低レベルAPI統合完了
- HTB, TBF, PRIO, FQ_CODEL qdisc全て対応済み
- U32フィルター、statistics収集対応

### 2. テスティング
**stretchr/testify v1.8.4** ✅
- アサーションライブラリ使用中
- 包括的テストスイート完成
- モック実装完了

### 3. ロギング
**uber-go/zap v1.26.0** ✅ **NEW**
- 構造化ログ完全実装
- 高パフォーマンス保証
- コンポーネント別レベル管理
- 開発・本番環境設定完備

### 4. データベース
**modernc.org/sqlite** ✅ **NEW**
- SQLite Event Store実装
- イベントソーシング永続化
- トランザクション対応
- スキーママイグレーション

### 5. 設定管理
**gopkg.in/yaml.v3** ✅
- YAML/JSON設定ファイルサポート
- 階層構造設定対応
- バリデーション機能完備

**encoding/json** (標準) ✅
- JSON設定サポート
- 構造化データ管理

## 開発ツール実装状況

### ビルド・リリース ✅
- **Makefile**: 簡素化されたビルドシステム
- **GoReleaser**: マルチプラットフォームビルド
- **Release Please**: 自動バージョン管理
- **golangci-lint**: GitHub Actions統合済み
- **go fmt**: 自動フォーマット設定済み
- **go vet**: 静的解析実行中
- **gosec**: セキュリティスキャン実装済み

### ドキュメント ✅
- **godoc**: 全public API文書化済み
- **README.md**: 包括的なプロジェクト概要
- **docs/**: 詳細な機能別ドキュメント
  - 構造化設定API
  - 優先度システム
  - ロギングシステム

### CI/CD ✅
- **GitHub Actions**: 完全自動化済み
  - CIワークフロー: テスト、リント、ビルド
  - リリースワークフロー: GoReleaser統合
  - Release Please: 自動バージョン管理
  - 複数Goバージョン対応（1.20, 1.21）
  - セキュリティスキャン統合

## システム要件

### 実行環境
- Linux kernel 3.10+ (TC機能に必要)
- root権限またはCAP_NET_ADMIN capability
- iproute2パッケージ（tcコマンド）

### 開発環境サポート ✅
- **Docker**: テスト環境コンテナ化済み
- **モック**: Linux環境非依存テスト完備
- **WSL2**: Windows開発者サポート

## パフォーマンス最適化実装

### ロギング最適化 ✅
- Zapのサンプリング機能
- 本番環境用設定テンプレート
- 非同期ログ出力オプション

### メモリ効率 ✅
- ポインタベース優先度管理
- 構造化設定の効率的パース
- ガベージコレクション配慮

### 並行性 ✅
- スレッドセーフなロガー実装
- コンテキスト対応API
- グローバル状態の最小化

## 技術的制約と対応

### 互換性対応 ✅
- Linux固有機能の抽象化
- Netlinkバージョン差異の吸収
- テスト環境での動作保証

### セキュリティ対応 ✅
- 入力バリデーション強化
- 構造化設定の安全な解析
- エラーハンドリングの統一化

## 実装済み技術アーキテクチャ

### Domain-Driven Design ✅
- Value Objects: Bandwidth, Device, Handle
- Entities: Class, Filter, Qdisc
- Aggregates: TrafficControl
- Events: ClassCreated, FilterAdded等

### CQRS パターン ✅
- Command handlers: HTB操作
- Query handlers: 設定取得
- Event sourcing: 設定変更履歴

### 関数型プログラミング ✅
- Result[T,E]型によるエラーハンドリング
- Immutable value objects
- Pure function重視

## 現在の依存関係 (go.mod)

```go
module github.com/rng999/traffic-control-go

go 1.21

require (
    github.com/stretchr/testify v1.8.4
    github.com/vishvananda/netlink v1.2.1
    github.com/vishvananda/netns v0.0.4
    go.uber.org/zap v1.26.0
    gopkg.in/yaml.v3 v3.0.1
    modernc.org/sqlite v1.27.0
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/dustin/go-humanize v1.0.1 // indirect
    github.com/google/uuid v1.3.0 // indirect
    github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
    github.com/mattn/go-isatty v0.0.16 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
    go.uber.org/multierr v1.10.0 // indirect
    golang.org/x/mod v0.3.0 // indirect
    golang.org/x/sys v0.15.0 // indirect
    golang.org/x/tools v0.0.0-20201124115921-2c860bdd6e78 // indirect
    golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
    lukechampine.com/uint128 v1.2.0 // indirect
    modernc.org/cc/v3 v3.40.0 // indirect
    modernc.org/ccgo/v3 v3.16.13 // indirect
    modernc.org/libc v1.24.1 // indirect
    modernc.org/mathutil v1.5.0 // indirect
    modernc.org/memory v1.6.0 // indirect
    modernc.org/opt v0.1.3 // indirect
    modernc.org/strutil v1.1.3 // indirect
    modernc.org/token v1.0.1 // indirect
)
```

## 将来の技術検討

### 次期実装候補
1. **NETEM qdisc** (高優先度)
   - ネットワークエミュレーション
   - レイテンシ、ジッター、パケットロス
   
2. **fw/flower filters** (中優先度)
   - ファイアウォールマークフィルター
   - 高度なパケットマッチング
   
3. **police/mirred actions** (中優先度)
   - レート制限アクション
   - パケットミラーリング

4. **eBPF統合** (低優先度)
   - カスタムパケット処理
   - パフォーマンス向上

### アーキテクチャ進化
- **パフォーマンス測定**: ベンチマークスイート
- **Web UI**: ブラウザベース管理インターフェース
- **Observability**: Prometheus/Grafana連携
- **Plugin システム**: カスタムqdisc/filter拡張

## 品質保証

### テストカバレッジ ✅
- Unit tests: 85%+
- Integration tests: 主要機能
- Example tests: 全サンプルコード

### 静的解析 ✅
- golangci-lint: 全種別チェック
- gosec: セキュリティスキャン
- go vet: 潜在的バグ検出

### ドキュメント品質 ✅
- API documentation: 100%
- Usage examples: 包括的
- Architecture guides: 詳細