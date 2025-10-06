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
	"github.com/team-swsd/circlehq/internal/core"
	"github.com/team-swsd/circlehq/internal/discord"
	"github.com/team-swsd/circlehq/internal/log"
	"github.com/team-swsd/circlehq/internal/server"
)

type ServeCmdOptions struct {
	Port string

	Debug   bool
	Address string
	NoAuth  bool
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
	flags.StringVar(&opts.Port, "port", "8000", "Port number for the rest server listening")
	flags.BoolVar(&opts.Debug, "debug", false, "Debug server mode")
	flags.BoolVar(&opts.NoAuth, "noauth", false, "bypass basic auth")

	return c
}

const (
	squareBaseURL     = "https://connect.squareupsandbox.com/"
	squareAccessToken = "EAAAl0L-AOHAQiNLA1UZvKXg0cjPYbp5jN_Wri_vK0wiF7XcA3as-QHCZACo5sdU"
	signatureKey      = "iOHasX0A3t7sQWmjnTozBw"
)

func runServe(opts *ServeCmdOptions) error {
	var logger *slog.Logger
	if opts.Debug {
		logger = log.NewLogger(os.Stdout, nil, slog.LevelDebug)
	} else {
		logger = log.NewLogger(os.Stdout, nil, slog.LevelInfo)
	}

	discordClient := discord.NewDiscordClient(logger, "webhook_url", 5*time.Second, "username", "avatar_url")

	squareClient := squareclient.NewClient(
		option.WithToken(squareAccessToken),
		option.WithBaseURL(squareBaseURL),
	)

	// Initialize the catalog
	catalog, err := catalog.NewCatalog(squareClient)
	if err != nil {
		logger.Error("failed to initialize catalog", "error", err)
		return err
	}
	logger.Info("catalog initialized", "items_count", len(catalog.Items))
	logger.Info("catalog items", "items", catalog.Items)

	broadcaster := broadcast.NewBroadcaster(logger)
	go broadcaster.Run()

	circleHQCore := core.NewCore(logger, catalog, discordClient, squareClient, broadcaster)
	circleHQService := server.NewCircleHQService(logger, circleHQCore, signatureKey)
	routerOpts := server.DefaultRouterOptions(server.RouterOptions{Logger: logger})
	handler := server.HandlerWithOptions(circleHQService, routerOpts)
	addr := net.JoinHostPort("", opts.Port)

	return server.NewHTTPServer(handler, addr, defaultShutdownSignal, defaultShutdownWaitTime, logger)
}
