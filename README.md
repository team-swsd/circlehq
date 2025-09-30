# CircleHQ

![logo](internal/server/static/circlehq.png)

now on beta

## 設計

### 大まかな流れ

1. 起動する
2. カタログ情報を取りに行く
3. API待ち受ける
4. Webhookでinventory update来たらDiscordに投げる
5. ブラウザでアクセスした場合はダッシュボードを表示
6. SSEでもクライアントに情報を渡す

### ダッシュボードで欲しいもの
- 本の情報
  - 名前
  - 通常販売残部数
  - 取り置き残部数
  - 退避部数

### 処理の主な流れ

- internal/cmdからcatalogを呼んで、client/squareを使ってカタログだけ取る。catalog内に定義した、catalog.SquareCatalog構造体をもらう
  - client/squareでカタログのjsonを取得、jsonをパース出来る構造体にunmarshal。
  - catalogパッケージで、catalog.SquareCatalog構造体にマッピング
  - ポインタを返す
- *catalog.SquareCatalogをserviceに渡す
- internal/cmdは、NewCoreをする
- internal/cmdは、NewService時にNewCoreを渡す
- Coreは呼ばれるたびにWebhookのペイロードとSquareCatalog内のアイテムIDと一致する名前を検索し、それともとにDiscordやSSEで名前や残部数を応答する

ブラウザでダッシュボードへアクセスされた場合
- アクセス時に、client/squareを使って、アイテム単位の在庫を問い合わせ、その情報を渡す
- SSEで更新を流す

ファイル構成
```
.
├── cmd
│   └── circlehq
│       └── main.go
├── internal
│   ├── catalog
│   │   └── catalog.go
│   ├── client
│   │   ├── discord
│   │   │   └── discord.go
│   │   └── square
│   │       └── square.go
│   ├── cmd
│   │   ├── root.go
│   │   └── serve.go
│   ├── core
│   │   └── core.go
│   ├── log
│   │   └── logger.go
│   ├── model
│   │   └── errors.go
│   └── server
│       ├── helper.go
│       ├── router.go
│       ├── server.gen.go
│       ├── server.go
│       ├── service.go
│       └── types.gen.go
```
