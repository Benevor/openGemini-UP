/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"openGemini-UP/pkg/install"
	"openGemini-UP/pkg/start"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start cluster",
	Long:  `Start an openGemini cluster based on configuration files and version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = StartCluster(ops)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("\nCheck the status of openGemini cluster\n")
		err = PatrolCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func StartCluster(ops install.ClusterOptions) error {
	starter := start.NewGeminiStarter(ops)
	defer starter.Close()

	if err := starter.PrepareForStart(); err != nil {
		return err
	}
	if err := starter.Start(); err != nil {
		return err
	}
	fmt.Printf("Successfully started the openGemini cluster with version : %s\n", ops.Version)
	return nil
}

func init() {
	ClusterCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("version", "v", "", "component version")
	startCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	startCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	startCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	startCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
