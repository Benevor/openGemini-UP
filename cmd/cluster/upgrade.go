package cluster

import (
	"fmt"
	"openGemini-UP/pkg/install"
	"openGemini-UP/pkg/start"
	"openGemini-UP/pkg/stop"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade cluster",
	Long:  `upgrade an openGemini cluster to the specified version`,
	Run: func(cmd *cobra.Command, args []string) {
		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("upgrade to cluster version: ", ops.Version)

		// destroy all services
		stop := stop.NewGeminiStop(ops, true)
		defer stop.Close()
		if err := stop.Prepare(); err != nil {
			fmt.Println(err)
			return
		}
		if err := stop.Run(); err != nil {
			fmt.Println(err)
		}

		// install cluster
		installer := install.NewGeminiInstaller(ops)
		defer installer.Close()

		if err := installer.PrepareForInstall(); err != nil {
			fmt.Println(err)
			return
		}
		if err := installer.Install(); err != nil {
			fmt.Println(err)
		}

		// start cluster
		starter := start.NewGeminiStarter(ops)
		defer starter.Close()

		if err := starter.PrepareForStart(); err != nil {
			fmt.Println(err)
			return
		}
		if err := starter.Start(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	ClusterCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringP("version", "v", "", "component name")
	upgradeCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	upgradeCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	upgradeCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	upgradeCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
