package broadcast

import (
	"log/slog"
)

// Broadcaster はSSEクライアントを管理し、メッセージをブロードキャストします。
type Broadcaster struct {
	logger *slog.Logger

	// ブロードキャストされるメッセージを送信するチャネル
	// バッファを持たせることで、Webhookハンドラがブロックされるのを防ぐ
	broadcast chan []byte

	// 新規クライアント登録用のチャネル
	registeringClients chan chan []byte

	// 切断したクライアント登録解除用のチャネル
	unregisteringClients chan chan []byte

	// 現在接続中のクライアントのセット
	clients map[chan []byte]bool
}

// NewBroadcaster は新しいBroadcasterインスタンスを作成します。
func NewBroadcaster(logger *slog.Logger) *Broadcaster {
	return &Broadcaster{
		logger:               logger,
		broadcast:            make(chan []byte, 10), // バッファサイズは適宜調整
		registeringClients:   make(chan chan []byte),
		unregisteringClients: make(chan chan []byte),
		clients:              make(map[chan []byte]bool),
	}
}

// Run はブロードキャスターのメインループを開始します。
// このメソッドはgoroutineとして実行する必要があります。
func (b *Broadcaster) Run() {
	for {
		select {
		case client := <-b.registeringClients:
			b.clients[client] = true
			b.logger.Info("Client registered", "total_clients", len(b.clients))

		case client := <-b.unregisteringClients:
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
				b.logger.Info("Client unregistered", "total_clients", len(b.clients))
			}

		case message := <-b.broadcast:
			b.logger.Info("Broadcasting message to clients", "num_clients", len(b.clients))
			for client := range b.clients {
				// 各クライアントへの送信がブロックしないように、selectとdefaultを使用
				select {
				case client <- message:
				default:
					// クライアントのチャネルが詰まっている場合、
					// そのクライアントを強制的に切断して登録解除する
					b.logger.Warn("Client channel is full. Closing connection.")
					delete(b.clients, client)
					close(client)
				}
			}
		}
	}
}

// Broadcast は接続中の全クライアントにメッセージを送信します。
// このメソッドは、Webhookハンドラなどから呼び出されます。
func (b *Broadcaster) Broadcast(message []byte) {
	b.broadcast <- message
}

// Register は新しいクライアントを登録します。
// SSEハンドラから呼び出されます。
func (b *Broadcaster) Register(client chan []byte) {
	b.registeringClients <- client
}

// Unregister はクライアントの登録を解除します。
// SSEハンドラでクライアントの切断を検知した際に呼び出されます。
func (b *Broadcaster) Unregister(client chan []byte) {
	b.unregisteringClients <- client
}
