# GitHub Automation Documentation

このドキュメントでは、traffic-control-goプロジェクトで設定されたGitHub Actionsによる自動化機能について説明します。

## 🔄 ワークフロー概要

### 1. PR Validation (`pr-validation.yml`)

**トリガー**: PR作成、編集、同期時

**機能**:
- Issue番号の参照チェック（必須）
- PRタイトルの形式チェック（推奨）
- PR説明文の質チェック（推奨）

**Issue番号チェック**:
以下のいずれかの形式でIssue番号を参照する必要があります：

```
✅ 自動クローズ形式（PRマージ時に自動でIssueがクローズされます）
- "Fixes #123"
- "Closes #123" 
- "Resolves #123"
- "Fix #123"
- "Close #123"
- "Resolve #123"

✅ 参照形式（Issueは自動クローズされません）
- "#123"
```

### 2. PR Merged Actions (`pr-merged.yml`)

**トリガー**: PRがマージまたはクローズされた時

**機能**:
- マージされたPRから自動クローズ対象Issueを抽出
- 該当Issueに完了コメントを追加
- PRが単純にクローズされた場合の通知

### 3. Issue Management (`issue-management.yml`)

**トリガー**: 
- CIワークフローが成功完了した時
- Issueにコメントが投稿された時

**機能**:
- CI成功後のIssue自動クローズ
- Issue管理コマンドの処理

**利用可能なコマンド**:
```
/close または /done - Issueを完了としてクローズ
/reopen - クローズされたIssueを再開
/help または /commands - ヘルプメッセージを表示
```

### 4. CI Updates (`ci.yml` - 更新済み)

**新機能**:
- PR成功時の通知
- Issue管理ワークフローとの連携

## 📋 Issue テンプレート

### Bug Report (`bug_report.yml`)
- 🐛 バグ報告用
- 再現手順、期待される動作、実際の動作などを構造化

### Feature Request (`feature_request.yml`)
- ✨ 新機能要求用
- 問題の明確化、提案解決策、優先度などを構造化

### Task/Improvement (`task.yml`)
- 📋 開発タスクや改善要求用
- カテゴリ分類、受け入れ条件、見積もりなどを構造化

## 📝 PR テンプレート

構造化されたPRテンプレートが以下を促進：
- Issue番号の必須参照
- 変更タイプの明確化
- テスト状況の確認
- レビューチェックリスト

## 🔄 自動化フロー

### 通常の開発フロー

1. **Issue作成**
   ```
   開発者がIssue作成 → 自動でラベル付与
   ```

2. **PR作成**
   ```
   PR作成 → Issue番号チェック → CI実行
   ```

3. **PR処理**
   ```
   PRマージ → Issue自動コメント → CI成功 → Issue自動クローズ
   ```

### エラーハンドリング

1. **Issue番号未参照**
   ```
   PR作成 → バリデーション失敗 → 修正が必要
   ```

2. **CI失敗**
   ```
   PRマージ → CI失敗 → Issueは自動クローズされない
   ```

3. **手動管理**
   ```
   Issue内で /close コマンド → 即座にクローズ
   ```

## ⚙️ 設定カスタマイズ

### Issue番号パターンの追加

`pr-validation.yml`の正規表現パターンを編集：

```javascript
const issuePatterns = [
  /(?:close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)\s+#\d+/gi,
  // 新しいパターンをここに追加
];
```

### コマンドの追加

`issue-management.yml`にコマンド処理を追加：

```javascript
// /newcommand コマンド
else if (comment === '/newcommand') {
  // コマンド処理をここに実装
}
```

### 通知のカスタマイズ

各ワークフローのコメント文面やSlack/Discord連携を追加可能。

## 🔧 トラブルシューティング

### よくある問題

1. **権限エラー**
   - GITHUB_TOKENの権限を確認
   - ワークフローファイルの権限設定を確認

2. **Issue番号が認識されない**
   - 正規表現パターンを確認
   - Issue番号の形式を確認

3. **自動クローズが動作しない**
   - CI完了を確認
   - PRとIssueの関連付けを確認

### デバッグ方法

1. **ワークフロー実行ログの確認**
   - Actions タブで詳細ログを確認

2. **Issue/PRの関連付け確認**
   - GitHub上でlinkage情報を確認

3. **権限の確認**
   - Repository Settings > Actions で権限を確認

## 📈 効果測定

この自動化により以下が改善されます：

- ✅ Issue-PR間のトレーサビリティ向上
- ✅ プロジェクト管理の自動化
- ✅ 開発者の負担軽減
- ✅ 品質管理の強化
- ✅ リリース準備の効率化

## 🔄 今後の拡張予定

- [ ] Slack/Discord通知連携
- [ ] 自動リリースノート生成
- [ ] パフォーマンス回帰検知
- [ ] 依存関係の自動更新
- [ ] セキュリティスキャンの強化