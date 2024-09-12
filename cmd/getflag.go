package cmd

import (
	"fmt"
	"os"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/lib/ssh"
	"github.com/GoToolSharing/htb-cli/lib/submit"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var getflagCmd = &cobra.Command{
	Use:   "getflag",
	Short: "Gets and submits flags from a SSH connection (Linux only). It will try to get the flag from the current SSH user.",
	Run: func(cmd *cobra.Command, args []string) {

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		private_key, _ := cmd.Flags().GetString("privkey")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")

		if host == "" {
			fmt.Println("error: invalid host specified")
			return
		}

		if username == "" && private_key == "" {
			fmt.Println("error: please specify an username/password pair or a private key location with -privkey.")
			return
		}

		connection, hostname, err := ssh.Connect(username, password, host, port, private_key)

		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}

		userFlag, err := ssh.GetFlag(connection)

		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}

		// Close the connection as we won't be using it anymore
		connection.Close()

		url, payload, err := ssh.BuildSubmitStuff(hostname, userFlag)

		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}

		message, err := submit.SubmitFlag(url, payload)

		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}

		fmt.Println(message)
		config.GlobalConfig.Logger.Info("Exit getflag command correctly")
	},
}

func init() {
	rootCmd.AddCommand(getflagCmd)
	getflagCmd.Flags().StringP("username", "u", "", "SSH username")
	getflagCmd.Flags().StringP("password", "p", "", "SSH password")
	getflagCmd.Flags().String("privkey", "", "User's private key location")
	getflagCmd.Flags().IntP("port", "P", 22, "(Optional) SSH Port (Default 22)")
	getflagCmd.Flags().StringP("host", "", "", "(Optional) SSH host")
}
