package cmd

import (
	"log/slog"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	squareclient "github.com/square/square-go-sdk/client"
	"github.com/square/square-go-sdk/option"
	"github.com/team-swsd/circlehq/internal/broadcast"
	"github.com/team-swsd/circlehq/internal/catalog"
	"github.com/team-swsd/circlehq/internal/config"
	"github.com/team-swsd/circlehq/internal/core"
	"github.com/team-swsd/circlehq/internal/discord"
	"github.com/team-swsd/circlehq/internal/log"
	"github.com/team-swsd/circlehq/internal/server"
)

type ServeCmdOptions struct {
	Debug      bool
	NoAuth     bool
	configPath string
}

var (
	defaultShutdownSignal = []os.Signal{
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	}

	defaultShutdownWaitTime = 30 * time.Second
)

func NewServeCmd() *cobra.Command {
	opts := &ServeCmdOptions{}

	c := &cobra.Command{
		Use:   "serve",
		Short: "Start a CircleHQ server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServe(opts); err != nil {
				// error log output
				os.Exit(1)
			}
		},
	}

	flags := c.Flags()
	flags.StringVarP(&opts.configPath, "config", "c", "", "config file path")
	flags.BoolVar(&opts.Debug, "debug", false, "Debug server mode")
	flags.BoolVar(&opts.NoAuth, "noauth", false, "bypass basic auth")

	return c
}

func runServe(opts *ServeCmdOptions) error {
	var logger *slog.Logger
	if opts.Debug {
		logger = log.NewLogger(os.Stdout, nil, slog.LevelDebug)
	} else {
		logger = log.NewLogger(os.Stdout, nil, slog.LevelInfo)
	}

	configs, err := config.LoadConfig(opts.configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return err
	}

	discordClient := discord.NewDiscordClient(logger, configs.Discord.WebhookURL, 5*time.Second, configs.Discord.Username, configs.Discord.AvatarURL)

	squareClient := squareclient.NewClient(
		option.WithToken(configs.Square.AccessToken),
		option.WithBaseURL(configs.Square.BaseURL),
	)

	// Initialize the catalog
	catalog, err := catalog.NewCatalog(logger, squareClient)
	if err != nil {
		logger.Error("failed to initialize catalog", "error", err)
		return err
	}
	logger.Info("catalog initialized", "items_count", len(catalog.Items))
	logger.Info("catalog items", "items", catalog.Items)

	broadcaster := broadcast.NewBroadcaster(logger)
	go broadcaster.Run()

	circleHQCore := core.NewCore(logger, catalog, discordClient, squareClient, broadcaster)
	circleHQService := server.NewCircleHQService(logger, circleHQCore, configs.Square.SignatureKey, configs.Spreadsheet.GoogleSpreadsheetURL)
	routerOpts := server.DefaultRouterOptions(server.RouterOptions{Logger: logger})
	handler := server.HandlerWithOptions(circleHQService, routerOpts)
	addr := net.JoinHostPort(configs.Server.ListenAddress, configs.Server.ListenPort)

	logger.Info("Initialized!")
	logger.Info("start server", "address", addr)
	return server.NewHTTPServer(handler, addr, defaultShutdownSignal, defaultShutdownWaitTime, logger)
}
