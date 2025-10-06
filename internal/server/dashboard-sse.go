package server

import (
	"fmt"
	"net/http"
)

// dashboard api
// (GET /api/dashboard-sse)
func (chqs *CircleHQService) GetDashboardData(w http.ResponseWriter, r *http.Request) {

	// SSEに必要なヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// レスポンスをフラッシュするためのFlusherを取得
	flusher, ok := w.(http.Flusher)
	if !ok {
		chqs.logger.Error("Streaming unsupported")
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// このクライアント専用のメッセージチャネルを作成
	messageChan := make(chan []byte)

	// Broadcasterにこのクライアントを登録
	chqs.core.Broadcaster().Register(messageChan)
	chqs.logger.Info("SSE client connected")

	// この関数が終了する際に、必ずクライアントの登録を解除する
	defer func() {
		chqs.core.Broadcaster().Unregister(messageChan)
		chqs.logger.Info("SSE client disconnected")
	}()

	// クライアントからの切断を検知するためのチャネル
	ctx := r.Context()
	done := ctx.Done()

	for {
		select {
		case <-done:
			// クライアントが接続を切断した
			return
		case message := <-messageChan:
			// Broadcasterからメッセージを受信した
			// SSEのデータフォーマットに従って書き込む
			// "data: <json-string>\n\n"
			fmt.Fprintf(w, "data: %s\n\n", message)
			// データをクライアントに即時送信
			flusher.Flush()
		}
	}
}
