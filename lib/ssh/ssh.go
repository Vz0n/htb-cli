package ssh

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/lib/utils"
	"golang.org/x/crypto/ssh"
)

func Connect(username, password, host string, port int, private_key string) (*ssh.Client, string, error) {

	var priv_key ssh.Signer
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if private_key != "" {

		file, err := os.OpenFile(private_key, os.O_RDONLY, 0755)
		key_bytes, err2 := io.ReadAll(file)

		if err != nil || err2 != nil {
			return nil, "", fmt.Errorf("errors while reading file: %s/%s", err, err2)
		}

		signer, err := ssh.ParsePrivateKey(key_bytes)

		if err != nil {
			return nil, "", fmt.Errorf("error parsing key: %s", err)
		}

		priv_key = signer
		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(priv_key),
		}

	} else {
		config.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	}

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)

	if err != nil {
		return nil, "", fmt.Errorf("error: %s", err)
	}

	// Get hostname of the machine at the same time, as it will be used later.
	hostname, _ := getHostname(connection)

	fmt.Printf("SSH connection established with machine: %s\n", hostname)
	return connection, hostname, nil
}

func GetFlag(connection *ssh.Client) (string, error) {
	session, err := connection.NewSession()
	if err != nil {
		return "", fmt.Errorf("error creating session: %s", err)
	}
	defer session.Close()

	if connection.User() == "root" {
		// Get the root flag if we are root
		out, err := session.CombinedOutput("cat root.txt")

		if err != nil {
			return "", fmt.Errorf("error executing command: %s", err)
		}

		flag := strings.ReplaceAll(string(out), "\n", "")

		if len(flag) != 32 {
			return "", fmt.Errorf("error getting root flag: %s", flag)
		}

		return flag, nil
	} else {
		out, err := session.CombinedOutput("cat user.txt")
		if err != nil {
			return "", fmt.Errorf("error executing command: %s", err)
		}

		flag := strings.ReplaceAll(string(out), "\n", "")

		if len(flag) != 32 {
			return "", fmt.Errorf("error getting user flag: %s", flag)
		}

		return flag, nil
	}
}

func getHostname(connection *ssh.Client) (string, error) {
	hostnameSession, err := connection.NewSession()

	if err != nil {
		return "", err
	}

	sessionOutput, err := hostnameSession.CombinedOutput("hostname")

	if err != nil {
		return "", err
	}

	hostnameSession.Close()
	return strings.ReplaceAll(string(sessionOutput), "\n", ""), nil
}

func BuildSubmitStuff(hostname string, userFlag string) (string, map[string]string, error) {

	var payload map[string]string
	var url string

	machineID, err := utils.SearchItemIDByName(hostname, "Machine")
	if err != nil {
		return "", nil, err
	}
	machineType, err := utils.GetMachineType(machineID)
	if err != nil {
		return "", nil, err
	}
	config.GlobalConfig.Logger.Debug(fmt.Sprintf("Machine Type: %s", machineType))

	// TODO: Check if they are still the same
	if machineType == "release" {
		url = config.BaseHackTheBoxAPIURL + "/arena/own"
		payload = map[string]string{
			"flag": userFlag,
		}
	} else {
		url = config.BaseHackTheBoxAPIURL + "/machine/own"
		payload = map[string]string{
			"id":   machineID,
			"flag": userFlag,
		}
	}

	return url, payload, nil
}
