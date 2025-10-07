package main

import (
	"syscall/js" // JavaScriptとの連携に必要
)

var (
	// パッケージレベルの変数として宣言し、複数の関数からアクセスできるようにする
	doc           js.Value
	tabButtons    js.Value
	contentFrames js.Value
)

// 指定されたタブをアクティブにし、対応するiframeを表示する関数
func activateTab(tabToActivate js.Value) {
	// data-target属性から対応するiframeのIDを取得
	targetID := tabToActivate.Get("dataset").Get("target").String()
	targetFrame := doc.Call("getElementById", targetID+"-frame")

	// すべてのタブボタンから'active'クラスを削除
	for i := 0; i < tabButtons.Length(); i++ {
		tabButtons.Index(i).Get("classList").Call("remove", "active")
	}
	// クリックされたタブに'active'クラスを追加
	tabToActivate.Get("classList").Call("add", "active")

	// すべてのiframeから'active'クラスを削除
	for i := 0; i < contentFrames.Length(); i++ {
		contentFrames.Index(i).Get("classList").Call("remove", "active")
	}
	// 対応するiframeに'active'クラスを追加
	if !targetFrame.IsUndefined() {
		targetFrame.Get("classList").Call("add", "active")
	}
}

// タブボタンがクリックされたときに呼ばれるコールバック関数
func tabClickHandler(this js.Value, args []js.Value) interface{} {
	// thisはクリックされた要素自身を指す
	activateTab(this)
	return nil
}

// リロードボタンがクリックされたときに呼ばれるコールバック関数
func reloadClickHandler(this js.Value, args []js.Value) interface{} {
	// 現在アクティブなiframeを取得
	activeFrame := doc.Call("querySelector", ".content-frame.active")

	// iframeが見つかった場合のみリロードを実行
	if !activeFrame.IsNull() && !activeFrame.IsUndefined() {
		// iframeのsrc属性を再設定することでリロードする
		activeFrame.Set("src", activeFrame.Get("src"))
	}
	return nil
}

func main() {
	// Goのプログラムが終了しないように、チャネルで待機する
	c := make(chan struct{}, 0)

	println("Go WebAssembly Initialized")

	// グローバルな `document` オブジェクトを取得
	doc = js.Global().Get("document")

	// 必要なDOM要素を取得
	tabButtons = doc.Call("querySelectorAll", ".tab-button")
	contentFrames = doc.Call("querySelectorAll", ".content-frame")
	reloadButton := doc.Call("getElementById", "tab-reload-button")

	// 各タブボタンにクリックイベントリスナーを登録
	for i := 0; i < tabButtons.Length(); i++ {
		button := tabButtons.Index(i)
		button.Call("addEventListener", "click", js.FuncOf(tabClickHandler))
	}

	// リロードボタンにクリックイベントリスナーを登録
	reloadButton.Call("addEventListener", "click", js.FuncOf(reloadClickHandler))

	// 初期表示時に最初のタブをアクティブにする
	if tabButtons.Length() > 0 {
		activateTab(tabButtons.Index(0))
	}

	// プログラムの終了を防ぐ
	<-c
}
