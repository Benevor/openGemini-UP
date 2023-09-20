/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"openGemini-UP/pkg/install"
	"openGemini-UP/pkg/status"

	"github.com/spf13/cobra"
)

// statusCmd
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "check cluster status",
	Long:  `Check the current running status of an openGemini cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = PatrolCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func PatrolCluster(ops install.ClusterOptions) error {
	patroler := status.NewGeminiStatusPatroller(ops)
	defer patroler.Close()

	if err := patroler.PrepareForPatrol(); err != nil {
		return err
	}
	if err := patroler.Patrol(); err != nil {
		return err
	}
	return nil
}

func init() {
	ClusterCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	statusCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	statusCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	statusCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
