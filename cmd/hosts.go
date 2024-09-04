package cmd

import (
	"fmt"

	"os"

	"github.com/GoToolSharing/htb-cli/lib/hosts"
	"github.com/spf13/cobra"
)

var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "Interact with hosts file",
	Long:  "Add or remove names of your hosts file.",
	Run:   coreHostsCmd,
}

func coreHostsCmd(cmd *cobra.Command, args []string) {

	ipParam, _ := cmd.Flags().GetString("ip")
	newHostname, _ := cmd.Flags().GetString("add")
	deleteHostname, _ := cmd.Flags().GetString("delete")
	// A bit of boolean boilerplate tho
	bothFlagsDefined := (newHostname != "" && deleteHostname != "")

	if ipParam == "" || (newHostname == "" && deleteHostname == "") {
		fmt.Println("Usage: htb-cli hosts [--add|--delete] <hostname> --ip <ip>")
		fmt.Println("Use \"htb-cli help hosts\" for more information.")
		return
	}

	if bothFlagsDefined {
		fmt.Println("You can't use both add and delete flag at the same time.")
		return
	}

	if newHostname != "" {
		fmt.Printf("Adding host %s to your hosts file...\n", newHostname)
		err := hosts.AddEntryToHosts(ipParam, newHostname)

		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}

		return
	} else {
		fmt.Printf("Removing host %s from your hosts file...\n", deleteHostname)
		err := hosts.RemoveEntryFromHosts(ipParam, deleteHostname)

		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}

		return
	}

}

func init() {
	rootCmd.AddCommand(hostsCmd)
	hostsCmd.Flags().StringP("add", "a", "", "Add a new entry")
	hostsCmd.Flags().StringP("ip", "i", "", "IP Address")
	hostsCmd.Flags().StringP("delete", "d", "", "Delete an entry")
}
