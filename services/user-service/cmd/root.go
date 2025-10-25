/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/wutthichod/sa-connext/shared/config"
)

type configContextKey struct{}

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "connext",
	Short: "A brief description of your application",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		_ = godotenv.Load("./.env")

		cfg, err := config.InitConfig()
		if err != nil {
			return err
		}
		cmd.SetContext(context.WithValue(cmd.Context(), configContextKey{}, cfg))

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(serveCmd, setupCmd)
}

func getConfigFromCmd(cmd *cobra.Command) (config.Config, error) {
	cfg, ok := cmd.Context().Value(configContextKey{}).(config.Config)
	if !ok {
		return nil, errors.New("config not found")
	}
	return cfg, nil
}
