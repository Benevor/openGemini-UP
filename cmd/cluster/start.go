/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"openGemini-UP/pkg/start"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start cluster",
	Long:  `Start an openGemini cluster based on configuration files and version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster start! ---------------")

		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		starter := start.NewGeminiStarter(ops)
		defer starter.Close()

		if err := starter.PrepareForStart(); err != nil {
			fmt.Println(err)
			return
		}
		if err := starter.Start(); err != nil {
			fmt.Println(err)
		}
		fmt.Println("--------------- Successfully completed cluster start! ---------------")
	},
}

func init() {
	clusterCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("version", "v", "", "component name")
	startCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	startCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	startCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	startCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
