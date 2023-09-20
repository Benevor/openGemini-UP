/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"openGemini-UP/pkg/install"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install cluster",
	Long:  `Install an openGemini cluster based on configuration files and version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = InstallCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func InstallCluster(ops install.ClusterOptions) error {
	installer := install.NewGeminiInstaller(ops)
	defer installer.Close()

	if err := installer.PrepareForInstall(); err != nil {
		return err
	}
	if err := installer.Install(); err != nil {
		return err
	}
	return nil
}

func init() {
	ClusterCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("version", "v", "", "component version")
	installCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	installCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	installCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	installCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
