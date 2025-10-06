package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"syscall/js"
	"time"

	"github.com/team-swsd/circlehq/internal/model"
)

// SSEで受け取るイベントデータの全体構造
type SSEEvent struct {
	Payload   model.Dashboard `json:"payload"`
	Timestamp int64           `json:"timestamp"`
	Type      string          `json:"type"`
}

// HTMLテンプレート
const dashboardHTMLBody = `
  <h1>Dashboard</h1>
  <div class="meta">
    アイテム数: {{ len .Items }}<br>
    最終更新日時: {{ .UpdatedAt }}
  </div>

  {{ if not .Items }}
    <p class="empty">アイテムはありません。</p>
  {{ else }}
    <div class="items">
      {{ range $i, $item := .Items }}
        <section class="item">
          <h2>{{ $item.Name }}</h2>
          <div class="item-id">ID: {{ $item.ID }}</div>
          <div class="variations-summary">バリエーション: {{ len $item.Variations }}</div>

          {{ if not $item.Variations }}
            <p class="empty">バリエーションはありません。</p>
          {{ else }}
            <div class="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>#</th>
                    <th>Name</th>
                    <th>Price</th>
                    <th>Sellable</th>
                    <th>Quantity</th>
                  </tr>
                </thead>
                <tbody>
                  {{ range $j, $v := .Variations }}
                    <tr>
                      <td>{{ add $j 1 }}</td>
                      <td>{{ $v.Name }}</td>
                      <td>{{ $v.Price }}</td>
                      <td>
                        {{ if $v.Sellable }}
                          <span class="tag sellable">SELLABLE</span>
                        {{ else }}
                          <span class="tag unsellable">UNSELLABLE</span>
                        {{ end }}
                      </td>
                      <td>{{ $v.Quantity }}</td>
                    </tr>
                  {{ end }}
                </tbody>
              </table>
            </div>
          {{ end }}
        </section>
      {{ end }}
    </div>
  {{ end }}

  <footer>Generated at {{ now }}</footer>
`

var (
	// パース済みのテンプレートを保持
	tmpl *template.Template
	// EventSourceのインスタンスを保持
	eventSource js.Value

	// イベントハンドラ用のjs.Funcをパッケージレベルの変数として保持
	onOpen    js.Func
	onMessage js.Func
	onError   js.Func
)

func main() {
	// 最初に一度だけテンプレートをパースします
	var err error
	tmpl, err = template.New("dashboard").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"now": func() string { return time.Now().Format("2006-01-02 15:04:05") },
	}).Parse(dashboardHTMLBody)

	if err != nil {
		log.Fatalf("テンプレートのパースに失敗しました: %v", err)
	}

	// イベントハンドラを初期化
	initEventHandlers()

	// JavaScript側からGoの関数を呼び出せるように登録
	js.Global().Set("startDashboardSSE", js.FuncOf(startDashboardSSE))

	// Goプログラムが終了しないように待機
	select {}
}

// initEventHandlers は、再利用するイベントハンドラを一度だけ初期化します。
func initEventHandlers() {
	// onopen: 接続成功時のハンドラ
	onOpen = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Println("SSE接続が確立しました。")
		return nil
	})

	// onmessage: メッセージ受信時のハンドラ
	onMessage = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Println("SSEメッセージを受信しました。") // ログ出力テスト
		if len(args) == 0 {
			return nil
		}
		event := args[0]
		data := event.Get("data").String()

		var sseEvent SSEEvent
		if err := json.Unmarshal([]byte(data), &sseEvent); err != nil {
			log.Printf("SSEデータのJSONデコードに失敗しました: %v, data: %s", err, data)
			return nil
		}

		// 受け取ったデータでHTMLをレンダリング
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, sseEvent.Payload); err != nil {
			log.Printf("テンプレートの実行に失敗しました: %v", err)
			return nil
		}

		// document.body の中身を、レンダリングしたHTMLで完全に置き換え
		js.Global().Get("document").Get("body").Set("innerHTML", buf.String())
		log.Println("ダッシュボードを更新しました。")

		return nil
	})

	// onerror: エラー発生時のハンドラ（接続切断など）
	onError = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Println("SSE接続でエラーが発生しました。5秒後に再接続します...")
		// 古い接続を閉じる
		if eventSource.Truthy() {
			eventSource.Call("close")
		}

		// 5秒後に再接続処理を再実行
		time.AfterFunc(5*time.Second, connectSSE)
		return nil
	})
}

// startDashboardSSEは、JavaScriptから呼び出されるエントリーポイントです。
func startDashboardSSE(this js.Value, args []js.Value) interface{} {
	go connectSSE()
	return nil
}

// connectSSEは、SSEエンドポイントへの接続とイベントハンドラの設定を行います。
func connectSSE() {
	log.Println("SSEエンドポイント /api/dashboard-sse に接続します...")

	esClass := js.Global().Get("EventSource")
	if !esClass.Truthy() {
		log.Println("このブラウザはEventSourceをサポートしていません。")
		return
	}
	eventSource = esClass.New("/api/dashboard-sse")

	// 初期化済みのイベントハンドラを設定
	eventSource.Set("onopen", onOpen)
	eventSource.Set("onmessage", onMessage)
	eventSource.Set("onerror", onError)

	// ★★★ connectSSE関数内での defer .Release() は行わない ★★★
}
