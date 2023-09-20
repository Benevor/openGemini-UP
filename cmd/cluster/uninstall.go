/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"openGemini-UP/pkg/install"
	"openGemini-UP/pkg/uninstall"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall cluster",
	Long:  `uninstall an openGemini cluster based on configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = UninstallCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func UninstallCluster(ops install.ClusterOptions) error {
	uninstaller := uninstall.NewGeminiUninstaller(ops)
	defer uninstaller.Close()

	if err := uninstaller.Prepare(); err != nil {
		return err
	}
	if err := uninstaller.Run(); err != nil {
		return err
	}
	fmt.Printf("Successfully uninstalled the openGemini cluster with version : %s\n", ops.Version)
	return nil
}

func init() {
	ClusterCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringP("version", "v", "", "component version")
	uninstallCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	uninstallCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	uninstallCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	uninstallCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
