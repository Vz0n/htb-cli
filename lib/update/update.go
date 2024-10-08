package update

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/lib/utils"
)

func Check(newVersion string) (string, error) {
	// Dev version
	config.GlobalConfig.Logger.Debug(fmt.Sprintf("config.Version: %s", config.Version))
	if config.Version == "dev" {
		config.GlobalConfig.Logger.Info("Development version detected")
		return "Development version", nil
	}

	// Just let this be more flexible.
	repo := "Vz0n/htb-cli"
	githubVersion := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	resp, err := utils.HTTPRequest(http.MethodGet, githubVersion, nil)
	if err != nil {
		return "", err
	}
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("error decoding JSON: %v", err)
	}
	config.GlobalConfig.Logger.Debug(fmt.Sprintf("release.TagName : %s", release.TagName))
	config.GlobalConfig.Logger.Debug(fmt.Sprintf("config.Version : %s", config.Version))
	var message string
	if release.TagName != config.Version && release.TagName != "" {
		message = fmt.Sprintf("A new update is now available ! (%s)\nUpdate with : go install github.com/%s@latest", release.TagName, repo)
	} else {
		message = fmt.Sprintf("You're up to date ! (%s)", config.Version)
	}

	return message, nil
}
