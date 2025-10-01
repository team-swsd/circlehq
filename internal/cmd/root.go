package cmd

import "github.com/spf13/cobra"

func NewCircleHQCmd(version string) *cobra.Command {
	c := &cobra.Command{
		Use:     "circlehq",
		Version: version,
		Short:   "Circle HQ",
	}
	c.AddCommand(NewServeCmd())
	return c
}
