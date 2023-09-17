/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/deploy"
	"openGemini-UP/util"

	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "manage cluster",
	Long:  `Manage openGemini clusters, including deploying, stopping, destroying, monitoring, etc.`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}

func getClusterOptions(cmd *cobra.Command) (deploy.ClusterOptions, error) {
	var ops deploy.ClusterOptions
	if version, _ := cmd.Flags().GetString("version"); version == "" {
		ops.Version = util.Download_default_version
	} else {
		ops.Version = version
	}
	if user, _ := cmd.Flags().GetString("user"); user == "" {
		return ops, fmt.Errorf("the user is required")
	} else {
		ops.User = user
	}
	password, _ := cmd.Flags().GetString("password")
	identity_file, _ := cmd.Flags().GetString("identity_file")
	if password == "" && identity_file == "" || password != "" && identity_file != "" {
		return ops, fmt.Errorf("the password and identity_file need one and only one")
	} else {
		ops.IdentityFile = identity_file
		ops.Password = password
		if identity_file != "" {
			ops.SshType = config.SSH_KEY
		} else {
			ops.SshType = config.SSH_PW
		}
	}
	return ops, nil
}
