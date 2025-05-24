package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/lib/utils"
	"github.com/GoToolSharing/htb-cli/lib/webhooks"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// setupSignalHandler configures a signal handler to stop the spinner and gracefully exit upon receiving specific signals.
func setupSignalHandler(s *spinner.Spinner) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		s.Stop()
		os.Exit(0)
	}()
}

// coreStartCmd starts a specified machine and returns a status message and any error encountered.
func coreStartCmd(machineChoosen string, machineID string) (string, error) {
	var err error
	if machineID == "" {
		machineID, err = utils.SearchItemIDByName(machineChoosen, "Machine")
		if err != nil {
			return "", err
		}

	}
	config.GlobalConfig.Logger.Info(fmt.Sprintf("Machine ID: %s", machineID))

	machineTypeChan := make(chan string)
	machineErrChan := make(chan error)
	userSubChan := make(chan string)
	userSubErrChan := make(chan error)

	go func() {
		machineType, err := utils.GetMachineType(machineID)
		machineTypeChan <- machineType
		machineErrChan <- err
	}()

	go func() {
		userSubscription, err := utils.GetUserSubscription()
		userSubChan <- userSubscription
		userSubErrChan <- err
	}()

	machineType := <-machineTypeChan
	err = <-machineErrChan
	if err != nil {
		return "", err
	}
	config.GlobalConfig.Logger.Info(fmt.Sprintf("Machine Type: %s", machineType))

	userSubscription := <-userSubChan
	err = <-userSubErrChan
	if err != nil {
		return "", err
	}

	config.GlobalConfig.Logger.Info(fmt.Sprintf("User subscription: %s", userSubscription))

	//isActive := utils.CheckVPN()
	//if !isActive {
	// 	isConfirmed := utils.AskConfirmation("No active VPN has been detected. Would you like to start it ?", batchParam)
	// 	if isConfirmed {
	//
	// 	}
	//}

	var url string
	var jsonData []byte

	url = config.BaseHackTheBoxAPIURL + "/vm/spawn"
	jsonData = []byte(fmt.Sprintf("{\"machine_id\":\"%s\"}", machineID))

	resp, err := utils.HtbRequest(http.MethodPost, url, jsonData)
	if err != nil {
		return "", err
	}

	message, ok := utils.ParseJsonMessage(resp, "message").(string)

	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	reg, _ := regexp.Compile("^Machine spawned!.*")

	if !reg.MatchString(message) {
		return message, nil
	}

	ip := "Undefined"
	startTime := time.Now()

	switch {
	case machineType == "release" || machineType == "seasonal":
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		setupSignalHandler(s)
		s.Suffix = " Waiting for the machine to start in order to fetch the IP address (this might take a while)."
		s.Start()
		defer s.Stop()
		timeout := time.After(5 * time.Minute)
	LoopRelease:
		for {
			select {
			case <-timeout:
				fmt.Println("Timeout (5 min) ! Exiting")
				s.Stop()
				return "", nil
			default:
				ip, err = utils.GetActiveMachineIP(machineType)
				if err != nil {
					return "", err
				}
				if ip != "" {
					s.Stop()
					break LoopRelease
				}
				time.Sleep(3 * time.Second)
			}
		}
	case userSubscription == "vip+":
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		setupSignalHandler(s)
		s.Suffix = " Waiting for the machine to start in order to fetch the IP address (this might take a while)."
		s.Start()
		defer s.Stop()
		timeout := time.After(5 * time.Minute)
	Loop:
		for {
			select {
			case <-timeout:
				fmt.Println("Timeout (5 min) ! Exiting")
				s.Stop()
				return "", nil
			default:
				ip, err = utils.GetActiveMachineIP(machineType)
				if err != nil {
					return "", err
				}
				if ip != "" {
					s.Stop()
					break Loop
				}
				time.Sleep(3 * time.Second)
			}
		}
	default:
		// Get IP address from active machine
		activeMachineData, err := utils.GetInformationsFromActiveMachine()

		if err != nil {
			return "", err
		}

		if activeMachineData["ip"] != nil {
			ip = activeMachineData["ip"].(string)
		} else {
			return "", errors.New("no ip has been returned, check server status")
		}
	}
	tts := time.Since(startTime)
	message = fmt.Sprintf("%s\nTarget: %s\nTime to spawn was %s !", message, ip, tts)
	return message, nil
}

// startCmd defines the "start" command which initiates the starting of a specified machine.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a machine",
	Long:  `Starts a Hackthebox machine specified in argument`,
	Run: func(cmd *cobra.Command, args []string) {
		config.GlobalConfig.Logger.Info("Start command executed")
		machineChoosen, err := cmd.Flags().GetString("machine")

		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}
		var machineID string
		if machineChoosen == "" {
			config.GlobalConfig.Logger.Info("Launching the latest released machine")
			machineID, err = utils.SearchLastReleaseArenaMachine()
			if err != nil {
				config.GlobalConfig.Logger.Error("", zap.Error(err))
				os.Exit(1)
			}
			config.GlobalConfig.Logger.Debug(fmt.Sprintf("Machine ID : %s", machineID))

		}
		output, err := coreStartCmd(machineChoosen, machineID)
		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}
		fmt.Println(output)
		err = webhooks.SendToDiscord("start", output)
		if err != nil {
			config.GlobalConfig.Logger.Error("", zap.Error(err))
			os.Exit(1)
		}
		config.GlobalConfig.Logger.Info("Exit start command correctly")
	},
}

// init adds the startCmd to rootCmd and sets flags for the "start" command.
func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("machine", "m", "", "Machine name")
}
