package config

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

type Settings struct {
	Verbose    bool
	ProxyParam string
	BatchParam bool
}

var GlobalConfig Settings

var ConfigFile map[string]string

var homeDir = os.Getenv("HOME")

var BaseDirectory = homeDir + "/.local/htb-cli"

const HostHackTheBox = "www.hackthebox.com"

const BaseHackTheBoxAPIURL = "https://" + HostHackTheBox + "/api/v4"

const Version = "v1.4.1"

// LoadConfig reads a configuration file from a specified filepath and returns a map of key-value pairs.
func LoadConfig(filepath string) (map[string]string, error) {
	config := make(map[string]string)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("incorrectly formatted line in configuration file : %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if err := validateConfig(key, value); err != nil {
			return nil, err
		}

		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig checks if the provided key-value pairs in the configuration are valid.
func validateConfig(key, value string) error {
	switch key {
	case "Logging", "Batch":
		if value != "True" && value != "False" {
			return fmt.Errorf("the value for '%s' must be 'True' or 'False', got : %s", key, value)
		}
	case "Proxy":
		if value != "False" && !isValidHTTPorHTTPSURL(value) {
			return fmt.Errorf("the URL for '%s' must be a valid URL starting with http or https, got : %s", key, value)
		}
	case "Discord":
		if value != "False" && !isValidDiscordWebhook(value) {
			return fmt.Errorf("the Discord webhook URL is invalid : %s", value)
		}
	}

	return nil
}

// isValidDiscordWebhook checks if a given URL is a valid Discord webhook.
func isValidDiscordWebhook(u string) bool {
	parsedURL, err := url.Parse(u)
	return err == nil && parsedURL.Scheme == "https" && strings.Contains(parsedURL.Host, "discord.com") && strings.Contains(parsedURL.Path, "/api/webhooks/")
}

// isValidHTTPorHTTPSURL checks if a given URL is valid and uses either the HTTP or HTTPS protocol.
func isValidHTTPorHTTPSURL(u string) bool {
	parsedURL, err := url.Parse(u)
	return err == nil && (parsedURL.Scheme == "http" || parsedURL.Scheme == "https")
}

// Init initializes the application by setting up necessary directories, creating a default configuration file if it doesn't exist, and loading the configuration.
func Init() error {
	if _, err := os.Stat(BaseDirectory); os.IsNotExist(err) {
		log.Printf("The \"%s\" folder does not exist, creation in progress...\n", BaseDirectory)
		err := os.MkdirAll(BaseDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("folder creation error: %s", err)
		}

		log.Printf("\"%s\" folder created successfully\n\n", BaseDirectory)
	}

	confFilePath := BaseDirectory + "/default.conf"
	if _, err := os.Stat(confFilePath); os.IsNotExist(err) {
		file, err := os.Create(confFilePath)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		configContent := `Discord = False`

		writer := bufio.NewWriter(file)
		_, err = writer.WriteString(configContent)
		if err != nil {
			return fmt.Errorf("error when writing to file: %v", err)
		}

		err = writer.Flush()
		if err != nil {
			return fmt.Errorf("error clearing buffer: %v", err)
		}

		log.Println("Configuration file created successfully.")
	}

	log.Println("Loading configuration file...")
	config, err := LoadConfig(BaseDirectory + "/default.conf")
	if err != nil {
		return fmt.Errorf("error loading configuration file : %v", err)
	}

	log.Println("Configuration successfully loaded :", config)
	ConfigFile = config
	return nil
}