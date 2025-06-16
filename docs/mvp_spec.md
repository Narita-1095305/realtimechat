# 📦 リアルタイムチャット MVP 要件定義

## 🎯 ゴール

複数のユーザがブラウザからログインし、固定チャットルーム「general」でリアルタイムにテキストを送受信できる。

## ✅ スコープ（MVP）

### 機能

* [x] サインアップ / ログイン（JWT 認証）
* [x] 固定チャンネル "general" のみを表示
* [x] チャット画面でのリアルタイム送受信（WebSocket）
* [x] メッセージ保存（PostgreSQL）
* [ ] チャンネル作成・削除（除外）
* [ ] ファイル添付、画像送信（除外）
* [ ] プロフィール編集（除外）

### 非機能

* [x] docker-compose でローカル環境を構築可能にする
* [x] JWT 秘密鍵、DB 接続情報などは .env で管理
* [x] ユーザデータとメッセージデータは初期マイグレーションで作成

## 🧱 技術スタック

| 分類       | 技術                                            | 備考                      |
| -------- | --------------------------------------------- | ----------------------- |
| 言語       | Go, TypeScript                                | -                       |
| バックエンド   | Gin, GORM, nhooyr/websocket                   | Gin で REST+WS、GORMでDB操作 |
| フロントエンド  | Next.js 15 (App Router), TailwindCSS, Zustand | UI構築と状態管理               |
| DB       | PostgreSQL 16                                 | -                       |
| リアルタイム通信 | Redis Pub/Sub                                 | WS用メッセージ中継（単一ノード）       |
| 状態管理     | Zustand, TanStack Query                       | SWRパターン                 |
| その他      | Docker, docker-compose                        | ローカル開発用                 |

## 🧭 画面構成（Next.js App Router 構成）

```
/login         ログイン画面（SignUpも兼用）
/channels      固定の "general" チャンネル表示
/chat/general  チャット画面（リアルタイム送受信）
```

## 🗂️ ディレクトリ構成（トップレベルのみ）

```
repo/
├─ backend/
│   ├─ cmd/server/main.go
│   ├─ internal/
│   │   ├─ handler/         # Gin ルーティング
│   │   ├─ websocket/       # Hub / Client
│   │   ├─ service/         # 業務ロジック
│   │   └─ repo/            # DBアクセス層
│   └─ migrations/          # 初期スキーマ
├─ frontend/
│   ├─ app/
│   │   ├─ login/page.tsx
│   │   ├─ channels/page.tsx
│   │   └─ chat/general/page.tsx
│   ├─ components/
│   └─ hooks/
├─ docker-compose.yml
└─ .env.example
```

## ✅ 実装順（開発用ガイド）

1. Gin バックエンドで SignUp/Login API を JWT 対応で実装
2. PostgreSQL に users / messages テーブルをマイグレーション
3. Redis Pub/Sub + WebSocket ハブを実装（1ルーム固定）
4. Next.js 側でログイン → general ルームへのメッセージ送信/受信を実装
5. Playwright などで簡易E2Eテスト（ログイン → メッセージ送受信）

## 📝 注意事項

* チャンネルは固定の "general" のみ。動的作成はスコープ外。
* メッセージは text のみ、画像やファイル添付は無し。
* JWTはHMAC署名、署名鍵は.envで渡す。リフレッシュトークンは無し。

---

この仕様をもとに開発を開始してください。最初は `backend/cmd/server` でサーバ起動まで進め、認証から段階的に構築してください。
