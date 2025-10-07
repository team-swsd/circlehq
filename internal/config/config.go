package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// --- 構造体定義 ---
// config.tomlファイル全体の構造に対応します。
type Config struct {
	Square      SquareConfig
	Discord     DiscordConfig
	Server      ServerConfig
	Spreadsheet SpreadsheetConfig
}

// [square]セクションの構造体
type SquareConfig struct {
	BaseURL      string `toml:"squareBaseURL"`
	AccessToken  string `toml:"squareAccessToken"`
	SignatureKey string `toml:"signatureKey"`
}

// [discord]セクションの構造体
type DiscordConfig struct {
	WebhookURL string `toml:"webhookURL"`
	Username   string `toml:"username"`
	AvatarURL  string `toml:"avatarURL"`
}

// [server]セクションの構造体
type ServerConfig struct {
	ListenAddress string `toml:"listenAddress"`
	ListenPort    string `toml:"listenPort"`
}

// [spreadsheet]セクションの構造体
type SpreadsheetConfig struct {
	GoogleSpreadsheetURL string `toml:"googleSpreadsheetURL"`
}

func LoadConfig(configPath string) (Config, error) {
	// --- ファイル読み込み & パース処理 ---

	// 実行ファイルのあるディレクトリを取得し、config.tomlへのパスを生成
	if configPath == "" {
		// 実行ファイル（バイナリ）のパスを取得

		exePath, err := os.Executable()
		if err != nil {
			return Config{}, fmt.Errorf("実行ファイルのパス取得に失敗しました: %v", err)
		}
		configPath = filepath.Join(filepath.Dir(exePath), "config.toml")
	}
	fmt.Printf("設定ファイルを読み込みます: %s\n", configPath)

	// 設定ファイルを読み込み、Config構造体にパース（デコード）する
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		// ファイルが存在しない、またはパースに失敗した場合
		return Config{}, fmt.Errorf("設定ファイルの読み込みまたはパースに失敗しました: %v", err)
	}

	// 読み込んだ内容の確認
	fmt.Println("--- 設定の読み込みが完了しました ---")
	fmt.Printf("Square Base URL: %s\n", config.Square.BaseURL)
	fmt.Printf("Discord Webhook URL: %s\n", config.Discord.WebhookURL)
	fmt.Printf("Server Listen Address: %s:%d\n", config.Server.ListenAddress, config.Server.ListenPort)
	fmt.Printf("Spreadsheet URL: %s\n", config.Spreadsheet.GoogleSpreadsheetURL)
	return config, nil
}

func MakeTemplate(configPath string) error {
	var buf = new(bytes.Buffer)

	var config = Config{
		Square: SquareConfig{
			BaseURL: "https://connect.squareupsandbox.com/",
		},
		Discord: DiscordConfig{
			WebhookURL: "https://discord.com/api/webhooks/",
		},
		Server: ServerConfig{
			ListenAddress: "0.0.0.0",
			ListenPort:    "8000",
		},
		Spreadsheet: SpreadsheetConfig{
			GoogleSpreadsheetURL: "https://",
		},
	}

	err := toml.NewEncoder(buf).Encode(config)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(buf.String())

	// 実行ファイルのあるディレクトリを取得し、config.tomlへのパスを生成
	if configPath == "" {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("実行ファイルのパス取得に失敗しました: %v", err)
		}
		configPath = filepath.Join(filepath.Dir(exePath), "config.toml")
	}
	fmt.Printf("設定ファイルを書き込み: %s\n", configPath)

	if err := os.WriteFile(configPath, buf.Bytes(), 0666); err != nil {
		return err
	}

	return nil
}
