package cluster

import (
	"fmt"
	"openGemini-UP/pkg/install"

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

		var old_version string
		if old_version, _ = cmd.Flags().GetString("old_version"); old_version == "" {
			fmt.Println("the old_version is required")
			return
		}

		err = UpgradeCluster(ops, old_version)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func UpgradeCluster(ops install.ClusterOptions, oldV string) error {
	oldOps := ops
	oldOps.Version = oldV

	// stop all services
	if err := StopCluster(oldOps); err != nil {
		return err
	}

	// uninstall openGeini
	if err := UninstallCluster(oldOps); err != nil {
		return err
	}

	// install new cluster
	if err := InstallCluster(ops); err != nil {
		return err
	}

	// start new cluster
	if err := StartCluster(ops); err != nil {
		return err
	}
	fmt.Printf("Successfully upgraded the openGemini cluster from %s to %s\n", oldV, ops.Version)
	return nil
}

func init() {
	ClusterCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringP("version", "v", "", "component version")
	upgradeCmd.Flags().StringP("old_version", "o", "", "component name")
	upgradeCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	upgradeCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	upgradeCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	upgradeCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
