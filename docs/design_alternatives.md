# Traffic Control API Design Alternatives

## 現在のチェーン式API
```go
tc.New("eth0").
    SetTotalBandwidth("100Mbps").
    CreateTrafficClass("Web").
        WithGuaranteedBandwidth("30Mbps").
        WithBurstableTo("60Mbps").
        ForPort(80, 443).
        And().
    Apply()
```

## 1. 宣言的/設定ベースAPI

### 利点
- 設定の再利用が容易
- ファイルから読み込み可能
- バージョン管理しやすい
- 全体像が見やすい

### 欠点
- 冗長になりがち
- 動的な設定が難しい

```go
// 構造体で定義
config := TrafficConfig{
    Device:    "eth0",
    Bandwidth: "1Gbps",
    Classes: []Class{
        {
            Name:       "critical",
            Guaranteed: "400Mbps",
            Maximum:    "800Mbps",
            Priority:   High,
            Rules: []Rule{
                {Match: "port:22", Action: "allow"},
                {Match: "port:443", Action: "allow"},
            },
        },
    },
}

tc.Apply(config)

// またはYAML/JSONから
tc.LoadConfig("traffic.yaml")
```

## 2. 自然言語風DSL

### 利点
- 非常に読みやすい
- 直感的
- ドキュメント不要なほど明確

### 欠点
- 実装が複雑
- 柔軟性に欠ける場合がある

```go
// 英語の文章のような構文
On("eth0").WithBandwidth("1Gbps").
    Reserve("400Mbps").For("Critical Services").
        CanBurstTo("800Mbps").
        WhenTrafficIs().OnPort(22, 443).
    Reserve("300Mbps").For("Web Traffic").
        WhenTrafficIs().OnPort(80).
    Apply()

// さらに自然に
traffic := On("eth0")
traffic.EnsureAtLeast("20Mbps").GoesTo("VoIP").When().Port(5060)
traffic.Limit("10Mbps").For("Torrents").OnPorts(6881..6889)
```

## 3. ルールベースAPI

### 利点
- iptablesユーザーに馴染みやすい
- ルールの追加・削除が簡単
- 条件と動作が明確

### 欠点
- 階層構造の表現が難しい

```go
tc := NewTrafficControl("eth0", "1Gbps")

// ルールを追加
tc.AddRule("port=22 -> bandwidth=50Mbps priority=high")
tc.AddRule("ip=192.168.1.100 -> bandwidth=100Mbps")
tc.AddRule("app=torrent -> bandwidth=10Mbps priority=low")

// または構造化
tc.Rule().
    Match(Port(22)).
    Action(Bandwidth("50Mbps"), Priority(High)).
    Add()

tc.Apply()
```

## 4. ビルダーパターン（現在の改良版）

### 利点
- 型安全
- IDEのサポートが良い
- エラーを早期発見

### 欠点
- やや冗長

```go
// より構造化されたビルダー
tc := TrafficControl.Builder().
    Device("eth0").
    TotalBandwidth("1Gbps").
    AddClass(
        Class.Builder().
            Name("critical").
            GuaranteedBandwidth("400Mbps").
            MaxBandwidth("800Mbps").
            AddFilter(Filter.Port(22, 443)).
            Build(),
    ).
    AddClass(
        Class.Builder().
            Name("normal").
            GuaranteedBandwidth("300Mbps").
            Build(),
    ).
    Build()

tc.Apply()
```

## 5. 関数型/パイプライン風API

### 利点
- 合成可能
- テストしやすい
- 純粋関数的

### 欠点
- Go初心者には難解

```go
// 関数の組み合わせ
config := Pipeline(
    Device("eth0"),
    Bandwidth("1Gbps"),
    Class("critical", 
        Guaranteed("400Mbps"),
        Burst("800Mbps"),
        Match(Ports(22, 443)),
    ),
    Class("normal",
        Guaranteed("300Mbps"),
        Match(Ports(80)),
    ),
)

Apply(config)

// または
eth0 := Device("eth0")
critical := Class("critical").With(Guaranteed("400Mbps"))
normal := Class("normal").With(Guaranteed("300Mbps"))

Apply(eth0, Bandwidth("1Gbps"), critical, normal)
```

## 6. グラフ/ツリーベースAPI

### 利点
- TC の階層構造を直接表現
- 視覚的に理解しやすい

### 欠点
- 初期学習コストが高い

```go
// ノードとエッジで構築
root := tc.Root("eth0", "1Gbps")

critical := root.AddChild("critical", "400Mbps")
critical.AddFilter(PortFilter(22, 443))

normal := root.AddChild("normal", "300Mbps")
normal.AddFilter(PortFilter(80))

voip := critical.AddChild("voip", "100Mbps")
video := critical.AddChild("video", "300Mbps")

root.Apply()
```

## 7. SQL風クエリAPI

### 利点
- 多くの開発者に馴染みがある
- 複雑な条件を表現しやすい

### 欠点
- TCの概念とのマッピングが必要

```go
tc.Query(`
    CREATE CLASS critical ON eth0
    WITH guaranteed_bandwidth = 400Mbps,
         max_bandwidth = 800Mbps
    WHERE port IN (22, 443)
`)

tc.Query(`
    CREATE CLASS normal ON eth0
    WITH guaranteed_bandwidth = 300Mbps
    WHERE port = 80
`)
```

## 8. タグベース/アノテーション風API

### 利点
- メタデータを付けやすい
- 後から検索・変更が容易

### 欠点
- 型安全性が低い

```go
tc := NewTrafficControl("eth0")

tc.Define("@critical @high-priority bandwidth:400Mbps burst:800Mbps port:22,443")
tc.Define("@normal @medium-priority bandwidth:300Mbps port:80")
tc.Define("@background @low-priority bandwidth:100Mbps app:torrent")

tc.Apply()

// 後から変更
tc.Update("@critical", "bandwidth:500Mbps")
tc.Remove("@background")
```

## 推奨: ハイブリッドアプローチ

実際には、用途に応じて複数のAPIスタイルを提供するのが良いでしょう：

```go
// 1. シンプルな用途向けの自然言語風
tc.On("eth0").
    Reserve("50Mbps").For("SSH").OnPort(22).
    Reserve("100Mbps").For("Web").OnPort(80, 443).
    Apply()

// 2. 複雑な設定向けの宣言的
config := TrafficConfig{...}
tc.ApplyConfig(config)

// 3. プログラマティックな制御向けのビルダー
builder := tc.NewBuilder("eth0")
for _, rule := range dynamicRules {
    builder.AddClass(rule.ToClass())
}
builder.Apply()
```

どのスタイルが最も使いやすいと思いますか？