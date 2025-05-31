# API Design - ヒューマンリーダブルなAPI設計

## 設計理念
Linux TCの複雑なコマンドを、人間が理解しやすい形に変換する。

## 従来のTC vs 新しいAPI

### 1. HTB設定の比較

**従来のTCコマンド:**
```bash
tc qdisc add dev eth0 root handle 1: htb default 10
tc class add dev eth0 parent 1: classid 1:1 htb rate 1000mbit ceil 1000mbit
tc class add dev eth0 parent 1:1 classid 1:10 htb rate 100mbit ceil 200mbit
tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 match ip dst 192.168.1.10/32 flowid 1:10
```

**新しいAPI:**
```go
// トラフィックコントローラーの作成
tc := trafficcontrol.New("eth0")

// 読みやすいメソッドチェーン
err := tc.
    SetTotalBandwidth("1Gbps").
    CreateTrafficClass("database").
        WithGuaranteedBandwidth("100Mbps").
        WithMaxBandwidth("200Mbps").
        ForDestination("192.168.1.10").
    Apply()
```

### 2. 優先度設定の比較

**従来のTCコマンド:**
```bash
tc qdisc add dev eth0 root handle 1: prio
tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dport 22 0xffff flowid 1:1
tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dport 80 0xffff flowid 1:2
```

**新しいAPI:**
```go
tc.CreatePriorityQueue().
    HighPriority().
        ForSSH().
        ForPort(22).
    MediumPriority().
        ForHTTP().
        ForHTTPS().
    LowPriority().
        ForEverythingElse().
    Apply()
```

## 主要なAPI設計パターン

### 1. Bandwidth（帯域幅）の表現
```go
// 人間が読みやすい単位
type Bandwidth struct {
    value float64
    unit  BandwidthUnit
}

// 文字列から作成
bandwidth := MustParseBandwidth("100Mbps")
bandwidth := MustParseBandwidth("1.5Gbps")
bandwidth := MustParseBandwidth("512Kbps")

// メソッドでも指定可能
tc.WithBandwidth(Mbps(100))
tc.WithBandwidth(Gbps(1.5))
```

### 2. Traffic Class（トラフィッククラス）の命名
```go
// 意味のある名前でクラスを作成
tc.CreateTrafficClass("voip").
    WithGuaranteedBandwidth("2Mbps").
    WithLowLatency()

tc.CreateTrafficClass("bulk-download").
    WithMaxBandwidth("500Mbps").
    WithFairQueuing()

// 自動的にhandle番号を管理（ユーザーは意識しない）
```

### 3. フィルタリングの簡潔な表現
```go
// IPアドレスベース
class.ForSource("192.168.1.0/24")
class.ForDestination("10.0.0.5")

// ポートベース
class.ForPort(80, 443)  // HTTP/HTTPS
class.ForPortRange(5000, 5100)

// アプリケーションベース
class.ForApplication("ssh")
class.ForApplication("http", "https")

// プロトコルベース
class.ForProtocol(TCP)
class.ForProtocol(UDP)
```

### 4. QoSプロファイル
```go
// 事前定義されたプロファイル
tc.ApplyProfile(profiles.VoIPOptimized())
tc.ApplyProfile(profiles.WebServerOptimized())
tc.ApplyProfile(profiles.HomeNetworkFair())

// カスタムプロファイル
profile := Profile{
    Name: "GameServerOptimized",
    Classes: []ClassConfig{
        {
            Name: "game-traffic",
            GuaranteedBandwidth: "50Mbps",
            Priority: High,
            Latency: UltraLow,
        },
    },
}
```

### 5. 監視とデバッグ
```go
// 現在の設定を人間が読める形で表示
config := tc.GetCurrentConfiguration()
fmt.Println(config.PrettyPrint())

// 統計情報
stats := tc.GetStatistics()
for _, class := range stats.Classes {
    fmt.Printf("Class %s: %s used, %d packets dropped\n",
        class.Name,
        class.CurrentBandwidth.HumanReadable(),
        class.DroppedPackets)
}

// 設定の検証
errors := tc.Validate()
for _, err := range errors {
    fmt.Printf("Warning: %s (suggestion: %s)\n", 
        err.Problem, 
        err.Solution)
}
```

## エラーハンドリング

### 分かりやすいエラーメッセージ
```go
// 悪い例
// Error: failed to add class 1:10: invalid argument

// 良い例
err := tc.Apply()
// Error: Cannot set guaranteed bandwidth (100Mbps) higher than parent's total bandwidth (50Mbps).
// Suggestion: Either reduce the guaranteed bandwidth or increase the parent's bandwidth limit.
```

### バリデーション
```go
// 設定適用前の検証
if err := tc.Validate(); err != nil {
    validationErr := err.(ValidationError)
    fmt.Printf("Configuration issue: %s\n", validationErr.Description)
    fmt.Printf("How to fix: %s\n", validationErr.Suggestion)
}
```

## 使用例

### 例1: ホームネットワークの公平な共有
```go
tc := trafficcontrol.New("eth0").
    SetTotalBandwidth("100Mbps")

// 各デバイスに公平に帯域を割り当て
tc.CreateTrafficClass("living-room-tv").
    WithGuaranteedBandwidth("20Mbps").
    WithBurstableTo("40Mbps").
    ForMACAddress("aa:bb:cc:dd:ee:01")

tc.CreateTrafficClass("kids-tablets").
    WithGuaranteedBandwidth("10Mbps").
    WithMaxBandwidth("20Mbps").
    ForMACAddresses("aa:bb:cc:dd:ee:02", "aa:bb:cc:dd:ee:03")

tc.CreateTrafficClass("work-laptop").
    WithGuaranteedBandwidth("30Mbps").
    WithBurstableTo("80Mbps").
    WithHighPriority().
    ForMACAddress("aa:bb:cc:dd:ee:04")

tc.Apply()
```

### 例2: Webサーバーの最適化
```go
tc := trafficcontrol.New("eth0").
    SetTotalBandwidth("10Gbps").
    EnableFairQueuing()

// Web トラフィックを優先
tc.HighPriority().
    ForPorts(80, 443).
    WithLowLatency()

// データベースレプリケーション
tc.CreateTrafficClass("db-replication").
    WithGuaranteedBandwidth("1Gbps").
    WithMaxBandwidth("2Gbps").
    ForPortRange(3306, 3310)

// バックアップ（低優先度）
tc.LowPriority().
    ForPort(873). // rsync
    WithMaxBandwidth("500Mbps")

tc.Apply()
```

## まとめ
このAPIデザインにより、Linux TCの強力な機能を維持しながら、設定の複雑さを大幅に削減し、意図が明確なコードを書くことができます。