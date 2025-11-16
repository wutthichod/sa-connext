package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wutthichod/sa-connext/services/user-service/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := getConfigFromCmd(cmd)
		if err != nil {
			return err
		}

		if err := server.InitServer(cfg); err != nil {
			return err
		}
		return nil
	},
}
