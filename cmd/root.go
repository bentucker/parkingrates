package cmd

import (
    "github.com/bentucker/parkingrates/server"
    "github.com/spf13/cobra"
)

var (
    cfgFile string

    rootCmd = &cobra.Command{
        Use: "parkingrates",
        Short: "An application to calculate parking rates for a given time" +
            " period, based on a pre-defined fee schedule.",
        Run: func(cmd *cobra.Command, args []string) {
            port, _ := cmd.Flags().GetInt("port")
            gwport, _ := cmd.Flags().GetInt("gwport")
            go server.RunServer(cfgFile, port)
            server.StartGateway(port, gwport)
        },
    }
)

// Execute executes the root command.
func Execute() {
    rootCmd.Execute()
}

func init() {
    rootCmd.PersistentFlags().Int("port",
        32884, "listen for requests on this port")
    rootCmd.PersistentFlags().Int("gwport",
        32885, "listen for REST requests on this port")
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config",
        "rates.json", "config file")
}
