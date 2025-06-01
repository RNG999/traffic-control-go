# イベントストアに格納されるデータ

イベントストアには、トラフィックコントロール設定の全ての変更履歴が**イベント**として保存されます。

## イベントの基本構造

全てのイベントには以下の共通フィールドがあります：

```go
type BaseEvent struct {
    aggregateID string    // 集約ID (例: "tc:eth0")
    eventType   string    // イベントタイプ (例: "HTBQdiscCreated")
    occurredAt  time.Time // 発生時刻
    version     int       // イベントバージョン番号
}
```

## 格納されるイベントの種類

### 1. Qdisc（キューイングディシプリン）関連イベント

#### HTBQdiscCreatedEvent
HTB qdiscが作成された時のイベント：
```go
{
    aggregateID: "tc:eth0",
    eventType: "HTBQdiscCreated",
    occurredAt: "2024-01-20T10:30:00Z",
    version: 1,
    DeviceName: "eth0",
    Handle: "1:",              // ハンドル (major:minor)
    DefaultClass: "1:999",     // デフォルトクラス
    R2Q: 10                    // Rate to Quantum比率
}
```

#### QdiscDeletedEvent
Qdiscが削除された時のイベント：
```go
{
    aggregateID: "tc:eth0",
    eventType: "QdiscDeleted",
    occurredAt: "2024-01-20T11:00:00Z",
    version: 5,
    DeviceName: "eth0",
    Handle: "1:"
}
```

### 2. Class（トラフィッククラス）関連イベント

#### HTBClassCreatedEvent
HTBクラスが作成された時のイベント：
```go
{
    aggregateID: "tc:eth0",
    eventType: "HTBClassCreated",
    occurredAt: "2024-01-20T10:31:00Z",
    version: 2,
    DeviceName: "eth0",
    Handle: "1:10",            // クラスハンドル
    Parent: "1:",              // 親ハンドル
    Name: "high-priority",     // クラス名
    Rate: "10mbps",            // 保証帯域
    Ceil: "100mbps",           // 最大帯域
    Burst: 1600,               // バーストサイズ
    Cburst: 1600               // Ceilバーストサイズ
}
```

#### ClassPriorityChangedEvent
クラスの優先度が変更された時のイベント：
```go
{
    aggregateID: "tc:eth0",
    eventType: "ClassPriorityChanged",
    occurredAt: "2024-01-20T10:45:00Z",
    version: 4,
    DeviceName: "eth0",
    Handle: "1:10",
    OldPriority: 4,
    NewPriority: 1
}
```

### 3. Filter（パケットフィルタ）関連イベント

#### FilterCreatedEvent
フィルタが作成された時のイベント：
```go
{
    aggregateID: "tc:eth0",
    eventType: "FilterCreated",
    occurredAt: "2024-01-20T10:32:00Z",
    version: 3,
    DeviceName: "eth0",
    Parent: "1:",              // 親ハンドル
    Priority: 100,             // フィルタ優先度
    Handle: "800::800",        // フィルタハンドル
    FlowID: "1:10",           // 転送先クラス
    Protocol: "ip",            // プロトコル
    Matches: [                 // マッチ条件
        {
            Type: "DestinationIP",
            Value: "192.168.1.100/32"
        },
        {
            Type: "DestinationPort",
            Value: "443"
        }
    ]
}
```

## イベントストアの特徴

### 1. イミュータブル（不変）
- 一度保存されたイベントは変更されません
- 設定の変更は新しいイベントとして追加されます

### 2. 完全な履歴
- 全ての変更履歴が時系列で保存されます
- 任意の時点の状態を再現できます

### 3. イベントソーシング
- 現在の状態は全てのイベントを順番に適用して再構築されます
- 例：eth0の現在の設定 = 全てのeth0関連イベントを再生

### 4. 監査証跡
- いつ、何が変更されたかの完全な記録
- トラブルシューティングや監査に有用

## 実装例

メモリイベントストアでは、以下のような構造でデータが保存されます：

```go
type MemoryEventStore struct {
    events map[string][]DomainEvent  // aggregateID → イベントリスト
    versions map[string]int          // aggregateID → 最新バージョン
}

// 例：
events["tc:eth0"] = [
    HTBQdiscCreatedEvent{...},      // version: 1
    HTBClassCreatedEvent{...},       // version: 2
    FilterCreatedEvent{...},         // version: 3
    ClassPriorityChangedEvent{...},  // version: 4
]
```

永続化実装（PostgreSQL等）では、イベントはJSONやバイナリ形式でデータベーステーブルに保存されます。