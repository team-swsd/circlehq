package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/team-swsd/circlehq/internal/config"
)

type ConfgenCmdOptions struct {
	configPath string
}

func NewConfgenCmd() *cobra.Command {
	opts := &ConfgenCmdOptions{}

	c := &cobra.Command{
		Use:   "confgen",
		Short: "Generate config toml file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runConfgen(opts); err != nil {
				// error log output
				os.Exit(1)
			}
		},
	}

	flags := c.Flags()
	flags.StringVarP(&opts.configPath, "config", "c", "", "config file path")

	return c
}

func runConfgen(opts *ConfgenCmdOptions) error {
	if err := config.MakeTemplate(opts.configPath); err != nil {
		return err
	}
	return nil
}
