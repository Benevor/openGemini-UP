/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/install"
	"openGemini-UP/util"

	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "manage cluster",
	Long:  `Manage openGemini clusters, including install, stop, destroy, monitor, etc.`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

func getClusterOptions(cmd *cobra.Command) (install.ClusterOptions, error) {
	var ops install.ClusterOptions
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
	key, _ := cmd.Flags().GetString("key")
	if password == "" && key == "" || password != "" && key != "" {
		return ops, fmt.Errorf("the password and key need one and only one")
	} else {
		ops.Key = key
		ops.Password = password
		if key != "" {
			ops.SshType = config.SSH_KEY
		} else {
			ops.SshType = config.SSH_PW
		}
	}

	if yPath, _ := cmd.Flags().GetString("yaml"); yPath == "" {
		return ops, fmt.Errorf("the path of cluster configuration file must be specified")
	} else {
		ops.YamlPath = yPath
	}
	return ops, nil
}
