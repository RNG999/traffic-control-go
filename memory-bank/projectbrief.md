# Traffic Control Library Project Brief

## 概要
Linux Traffic Control (TC) のためのGoライブラリを開発する。このライブラリは、Linux TCの複雑な機能をGoから簡単に利用できるようにすることを目的とする。

## プロジェクトスコープ

### 実装対象：Linux Traffic Control (TC)
docs/traffic-control.mdで詳細に説明されているLinux TCシステムをGoで抽象化する：

1. **Queueing Disciplines (Qdiscs)**
   - Classless Qdiscs: pfifo_fast, TBF, SFQ
   - Classful Qdiscs: HTB, PRIO, CBQ, HFSC
   - Modern AQM: FQ_CODEL, CAKE

2. **Classes**
   - HTBクラス階層の管理
   - 帯域幅の保証（rate）と上限（ceil）
   - バーストサイズの制御

3. **Filters**
   - u32分類器
   - IPアドレス、ポート、プロトコルによるマッチング
   - firewall markとの統合

4. **Actions**
   - police: レート制限
   - drop: パケット破棄
   - mirred: パケットミラーリング

### ライブラリの目標
1. **ヒューマンリーダブル**: 直感的で理解しやすいAPI設計
2. **型安全性**: Goの型システムを活用し、設定ミスを防ぐ
3. **イミュータブル**: Event Sourcingパターンで設定変更を管理
4. **テスタブル**: ユニットテストとインテグレーションテストの充実
5. **使いやすさ**: 複雑なTC設定を簡潔なAPIで提供

### API設計方針
- **Fluent Interface**: メソッドチェーンで読みやすく
- **明確な単位**: "1Gbps", "500Mbps"のような人間が理解しやすい表記
- **分かりやすい名前**: `major:minor`ではなく意味のある名前
- **デフォルト値**: 一般的な設定は自動化
- **エラーメッセージ**: 問題と解決方法を明確に

## アーキテクチャ方針
- CQRS: TC設定の読み取りと変更を分離
- Event Sourcing: 設定変更履歴を保持
- DDD: QdiscEntity, ClassAggregate, FilterValueObjectなど
- 関数型プログラミング: Result[T,E]型でエラーハンドリング
- 型駆動開発: 型から設計を始める

## 技術スタック
- 言語: Go
- TCとの通信: netlink (vishvananda/netlink)
- テスト: 標準のtestingパッケージ + testify
- イベントストア: 初期実装はメモリ内、後に永続化対応

## 主要なユースケース
1. 帯域幅制限（rate limiting）
2. QoS実装（VoIP優先など）
3. 公平なキュー管理（fair queuing）
4. バッファブロート対策（AQM）

## 次のステップ
1. ドメインモデルの詳細設計
2. netlinkライブラリの調査と選定
3. 基本的なHTB操作のプロトタイプ実装