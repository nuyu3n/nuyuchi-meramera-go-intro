# ぬゆちのメラメラGo入門 完成版のコード

Zennの本「[ぬゆちのメラメラGo入門](https://zenn.dev/nuyuchi/books/nuyuchi-meramera-go-intro)」で、1冊かけて育てた課題の締切管理プログラム `kadai` の完成版です。チャプター29を終えた時点のコードが入っています。

途中で分からなくなったときに、自分のコードと見比べるために使ってください。

## 動かし方

Go 1.27.4 で動作を確かめています。

```sh
go run . list      # 締切が近い順に一覧
go run . summary   # 状態ごとの件数
go run . soon      # 3日以内の課題
go run . remind    # 通知をまとめて送る（の真似）
go run . serve     # APIサーバー（http://localhost:8080）
```

課題は `tasks.txt` から読み込みます。

APIサーバーは、次のように試せます。

```sh
curl localhost:8080/tasks
curl -X POST localhost:8080/tasks -H "Content-Type: application/json" --data-binary @new.json
curl -X POST localhost:8080/tasks/3/done
curl "localhost:8080/tasks?pending=true"
```

## テスト

```sh
go test ./...
```

## ファイル

| ファイル | 内容 | 主なチャプター |
|---|---|---|
| `main.go` | コマンド（`list`・`summary`・`soon`・`remind`・`serve`） | 12, 24〜26, 29 |
| `server.go` | APIのハンドラ | 27〜29 |
| `store.go` | 課題を持っておく所（`Mutex` で守る） | 26, 29 |
| `server_test.go` | APIのテスト | 29 |
| `task/task.go` | 課題の型と、読み込み・表示・絞り込み | 14〜20, 29 |
| `task/task_test.go` | `task` パッケージのテスト | 21, 22 |
| `tasks.txt` | 課題のデータ | 9 |
| `new.json` | APIで追加する課題の例 | 29 |

練習問題の解答例は入っていません。本の中の解答例を見てください。

## ライセンス

[MIT License](LICENSE)
