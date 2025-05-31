# Architecture Overview - 全体アーキテクチャ設計

## システム全体像

```
┌─────────────────────────────────────────────────────────────┐
│                        Web UI                                │
│           (React/Vue.js Dashboard)                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Control Plane                              │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   REST API  │  │  gRPC Server │  │  WebSocket Hub   │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           Configuration Manager                      │   │
│  │    (CQRS Command/Query Handlers)                   │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Event Store                            │   │
│  │         (PostgreSQL + Kafka)                        │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Data Plane                              │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ TC Agent    │  │ K8s Sidecar  │  │  Proxy Server    │  │
│  │ (DaemonSet) │  │  Container   │  │  (iptables/eBPF)│  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
│         │                │                    │              │
│         └────────────────┴────────────────────┘              │
│                          │                                   │
│                          ▼                                   │
│              ┌──────────────────────┐                       │
│              │  TC Library (Core)   │                       │
│              │  - HTB, FQ_CODEL    │                       │
│              │  - Netlink API      │                       │
│              └──────────────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## コンポーネント詳細

### 1. Control Plane（制御プレーン）

#### Web UI
- **技術**: React or Vue.js + TypeScript
- **機能**:
  - トラフィック制御ルールの視覚的設定
  - リアルタイムモニタリングダッシュボード
  - マルチテナント管理
  - アラート設定

#### API Gateway
- **REST API**: 外部統合用
- **gRPC**: エージェント通信用（高効率）
- **WebSocket**: リアルタイム更新用

#### Configuration Manager
- **CQRS実装**: 設定の読み書き分離
- **Event Sourcing**: 全変更履歴の保持
- **Validation**: 設定の整合性チェック

### 2. Data Plane（データプレーン）

#### TC Agent
- **デプロイ方式**: Kubernetes DaemonSet
- **責務**:
  - Control Planeからの指示受信
  - ローカルTC設定の適用
  - メトリクス収集と送信
  - ヘルスチェック

#### Kubernetes Sidecar
- **実装方式**: サイドカーコンテナ
- **機能**:
  - Pod単位のトラフィック制御
  - 透過プロキシ（iptables/eBPF）
  - Service Mesh統合（Istio/Linkerd）

#### Proxy Server
- **トラフィック処理**:
  - L4/L7プロキシ機能
  - トラフィック分類
  - TC設定の動的適用

### 3. Core Library（コアライブラリ）

現在開発中のtrafficcontrol-goライブラリ：
- ヒューマンリーダブルAPI
- Linux TC抽象化
- イベント駆動アーキテクチャ

## デプロイメントパターン

### パターン1: エンタープライズ on-premise
```yaml
# Kubernetes DaemonSetとして展開
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: tc-agent
spec:
  template:
    spec:
      containers:
      - name: tc-agent
        image: trafficcontrol/agent:latest
        securityContext:
          capabilities:
            add: ["NET_ADMIN"]
```

### パターン2: SaaS マルチテナント
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Tenant A   │     │  Tenant B   │     │  Tenant C   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┴───────────────────┘
                           │
                    ┌──────▼──────┐
                    │Control Plane│
                    │(Multi-tenant)│
                    └──────┬──────┘
                           │
       ┌───────────────────┼───────────────────┐
       ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│Agent Cluster A│    │Agent Cluster B│    │Agent Cluster C│
└──────────────┘    └──────────────┘    └──────────────┘
```

### パターン3: サイドカープロキシ
```yaml
# Podにサイドカーとして注入
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
  - name: tc-sidecar
    image: trafficcontrol/sidecar:latest
    securityContext:
      capabilities:
        add: ["NET_ADMIN"]
```

## セキュリティ設計

### 認証・認可
- **mTLS**: エージェント⇔コントロールプレーン通信
- **JWT/OAuth2**: WebUI認証
- **RBAC**: 細かい権限制御

### ネットワークセキュリティ
- **暗号化通信**: TLS 1.3
- **ネットワーク分離**: Control/Data Plane分離
- **監査ログ**: 全操作の記録

## スケーラビリティ設計

### 水平スケーリング
- Control Plane: Kubernetes HPA
- Event Store: Kafka partitioning
- Agent: ノード数に応じて自動スケール

### パフォーマンス目標
- エージェント数: 10,000+
- 設定更新レイテンシ: < 1秒
- メトリクス収集間隔: 1秒

## 統合ポイント

### Kubernetes統合
- CRD (Custom Resource Definition)
- Admission Webhook
- Operator pattern

### Service Mesh統合
- Istio: EnvoyFilter
- Linkerd: Traffic Policy
- Consul Connect: Intentions

### 監視システム統合
- Prometheus: メトリクスエクスポート
- Grafana: ダッシュボード
- ELK Stack: ログ収集

## ロードマップ

### Phase 1: Core Library (現在)
- [x] ヒューマンリーダブルAPI
- [ ] 基本的なTC操作
- [ ] イベントソーシング実装

### Phase 2: Agent Development
- [ ] gRPCクライアント実装
- [ ] メトリクス収集
- [ ] 設定同期メカニズム

### Phase 3: Control Plane
- [ ] REST/gRPC API
- [ ] WebUI基本機能
- [ ] マルチテナント対応

### Phase 4: Production Ready
- [ ] HA構成
- [ ] 自動スケーリング
- [ ] エンタープライズ機能