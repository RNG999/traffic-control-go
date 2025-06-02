# System Patterns - アーキテクチャとパターン

## 全体アーキテクチャ

### レイヤー構造
```
┌─────────────────────────────────────────┐
│         Application Layer               │
│  (CLI tools, monitoring services)       │
├─────────────────────────────────────────┤
│         Public API Layer                │
│  (Simple, type-safe interfaces)         │
├─────────────────────────────────────────┤
│         Domain Layer                    │
│  (Business logic, CQRS handlers)        │
├─────────────────────────────────────────┤
│       Infrastructure Layer              │
│  (Netlink, Event Store, Persistence)    │
└─────────────────────────────────────────┘
```

## CQRS実装パターン

### コマンド側
```go
// Command
type CreateHTBQdiscCommand struct {
    DeviceName  DeviceName
    Handle      QdiscHandle
    DefaultClass ClassID
}

// Command Handler
type CreateHTBQdiscHandler struct {
    netlinkAdapter NetlinkAdapter
    eventStore     EventStore
}

// Command Result
type Result[T any] struct {
    value T
    err   error
}
```

### クエリ側
```go
// Query
type GetQdiscByDeviceQuery struct {
    DeviceName DeviceName
}

// Query Handler
type GetQdiscByDeviceHandler struct {
    readModel QdiscReadModel
}

// Read Model
type QdiscReadModel interface {
    GetByDevice(DeviceName) ([]QdiscView, error)
}
```

## Event Sourcing パターン

### イベント定義
```go
// Base Event
type DomainEvent interface {
    AggregateID() string
    EventType() string
    OccurredAt() time.Time
    Version() int
}

// Specific Events
type HTBQdiscCreatedEvent struct {
    DeviceName   DeviceName
    Handle       QdiscHandle
    DefaultClass ClassID
    Timestamp    time.Time
}

type HTBClassAddedEvent struct {
    ParentHandle Handle
    ClassID      ClassID
    Rate         Bandwidth
    Ceil         Bandwidth
    Timestamp    time.Time
}
```

### イベントストア
```go
type EventStore interface {
    Save(events []DomainEvent) error
    GetEvents(aggregateID string, fromVersion int) ([]DomainEvent, error)
    GetAllEvents() ([]DomainEvent, error)
}

// 実装済み
- MemoryEventStore: インメモリ実装（テスト用）
- SQLiteEventStore: 永続化実装（本番用）
```

## DDD ドメインモデル

### Value Objects
```go
// 帯域幅を表すValue Object
type Bandwidth struct {
    value uint64 // bits per second
    unit  BandwidthUnit
}

// ハンドルを表すValue Object  
type Handle struct {
    major uint16
    minor uint16
}

// デバイス名を表すValue Object
type DeviceName string
```

### Entities
```go
// Qdisc Entity
type QdiscEntity struct {
    id         QdiscID
    handle     Handle
    deviceName DeviceName
    qdiscType  QdiscType
    parent     *Handle
}

// Class Entity
type ClassEntity struct {
    id       ClassID
    handle   Handle
    parent   Handle
    rate     Bandwidth
    ceil     Bandwidth
    burst    Burst
}
```

### Aggregates
```go
// Traffic Control Configuration Aggregate
type TrafficControlAggregate struct {
    deviceName DeviceName
    qdiscs     []QdiscEntity
    classes    []ClassEntity
    filters    []FilterEntity
    version    int
    events     []DomainEvent
}

// Aggregate Methods
func (tc *TrafficControlAggregate) AddHTBQdisc(handle Handle, defaultClass ClassID) error
func (tc *TrafficControlAggregate) AddHTBClass(parent Handle, classID ClassID, rate, ceil Bandwidth) error
func (tc *TrafficControlAggregate) GetUncommittedEvents() []DomainEvent
```

## 関数型パターン

### Result型によるエラーハンドリング
```go
type Result[T any] struct {
    value T
    err   error
}

func (r Result[T]) IsSuccess() bool
func (r Result[T]) Map(f func(T) T) Result[T]
func (r Result[T]) FlatMap(f func(T) Result[T]) Result[T]
func (r Result[T]) Match(onSuccess func(T), onError func(error))
```

### Maybe型によるnull安全性
```go
type Maybe[T any] struct {
    value *T
}

func Some[T any](value T) Maybe[T]
func None[T any]() Maybe[T]
func (m Maybe[T]) Map(f func(T) T) Maybe[T]
func (m Maybe[T]) OrElse(defaultValue T) T
```

## インフラストラクチャパターン

### Netlink Adapter
```go
type NetlinkAdapter interface {
    // Qdisc操作
    AddQdisc(qdisc QdiscConfig) Result[Unit]
    DeleteQdisc(handle Handle, device DeviceName) Result[Unit]
    GetQdiscs(device DeviceName) Result[[]QdiscInfo]
    
    // Class操作
    AddClass(class ClassConfig) Result[Unit]
    DeleteClass(handle Handle, device DeviceName) Result[Unit]
    GetClasses(device DeviceName) Result[[]ClassInfo]
    
    // Filter操作
    AddFilter(filter FilterConfig) Result[Unit]
    DeleteFilter(handle Handle, device DeviceName) Result[Unit]
    
    // Statistics
    GetStatistics(handle Handle, device DeviceName) Result[Statistics]
}

// 実装済みQdiscタイプ
- HTB (Hierarchical Token Bucket)
- TBF (Token Bucket Filter)
- PRIO (Priority Scheduler)
- FQ_CODEL (Fair Queue Controlled Delay)
```

### Repository Pattern
```go
type TrafficControlRepository interface {
    Save(aggregate TrafficControlAggregate) error
    Load(deviceName DeviceName) (*TrafficControlAggregate, error)
}
```

## テスト戦略

### 単体テスト
- ドメインロジックのテスト（netlinkなし）
- Value Objectsの不変性テスト
- イベントの生成と適用のテスト

### 統合テスト
- Netlink adapterのモック/スタブ
- イベントストアのインメモリ実装
- エンドツーエンドのコマンド/クエリフロー

### システムテスト
- 実際のLinux環境での動作確認
- Docker/VMでの隔離環境テスト
- パフォーマンステスト