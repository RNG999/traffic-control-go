# SQLite イベントストア実装

## 概要

SQLiteを使用した永続的なイベントストアの実装です。PostgreSQLよりもシンプルで、組み込みやすいのが特徴です。

## 特徴

- **軽量**: 外部DBサーバー不要
- **ゼロ設定**: ファイルパスを指定するだけ
- **トランザクション対応**: ACID特性を保証
- **楽観的同時実行制御**: バージョンチェックによる競合検出

## データベース構造

### eventsテーブル
```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    aggregate_id TEXT NOT NULL,         -- 集約ID (例: "tc:eth0")
    event_type TEXT NOT NULL,           -- イベントタイプ
    event_data TEXT NOT NULL,           -- JSONシリアライズされたイベントデータ
    event_version INTEGER NOT NULL,     -- イベントバージョン
    occurred_at TIMESTAMP NOT NULL,     -- イベント発生時刻
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### snapshotsテーブル（将来の拡張用）
```sql
CREATE TABLE snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    aggregate_id TEXT NOT NULL UNIQUE,
    snapshot_data TEXT NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 使用方法

### 1. 作成
```go
// ファイルベースのSQLite
store, err := eventstore.NewSQLiteEventStoreWithContext("./events.db")

// インメモリSQLite
store, err := eventstore.NewSQLiteEventStoreWithContext(":memory:")
```

### 2. アグリゲートの保存
```go
aggregate := aggregates.NewTrafficControlAggregate(device)
// ... 操作を実行 ...

err := store.SaveAggregate(ctx, aggregate)
```

### 3. アグリゲートの読み込み
```go
aggregate := aggregates.NewTrafficControlAggregate(device)
err := store.Load(ctx, aggregateID, aggregate)
```

## 格納されるデータ例

```json
{
  "aggregateID": "tc:eth0",
  "eventType": "HTBQdiscCreated",
  "version": 1,
  "occurredAt": "2024-01-20T10:30:00Z",
  "DeviceName": "eth0",
  "Handle": "1:",
  "DefaultClass": "1:999"
}
```

## メリット

1. **シンプル**: 設定や管理が簡単
2. **ポータブル**: 単一ファイルでバックアップ/移行が容易
3. **高速**: ローカルファイルアクセスで低レイテンシ
4. **信頼性**: SQLiteの実績ある堅牢性

## 制限事項

1. **同時接続**: 書き込みは1つずつ（読み込みは並行可能）
2. **スケール**: 単一マシンでの使用に適している
3. **レプリケーション**: 標準機能では提供されない

## 本番環境での考慮事項

- 定期的なバックアップ（`VACUUM`コマンドの実行）
- ファイルシステムの権限設定
- ディスク容量の監視
- WAL（Write-Ahead Logging）モードの有効化を検討