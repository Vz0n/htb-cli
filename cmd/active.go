package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"net/http"

	"github.com/QU35T-code/htb-cli/utils"
	"github.com/kyokomi/emoji/v2"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Catalogue of active machines",
	Long:  "This command serves to generate a detailed summary of the currently active machines, providing pertinent information for each.",
	Run: func(cmd *cobra.Command, args []string) {
		url := "https://www.hackthebox.com/api/v4/machine/list"
		resp, err := utils.HtbRequest(http.MethodGet, url, proxyParam, nil)
		if err != nil {
			log.Fatal(err)
		}
		info := utils.ParseJsonMessage(resp, "info")
		log.Println(info)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "Name\tOS\tDifficulty\tUser Owns\tSystem Owns\tStars\tStatus\tRelease")

		// red_color := color.New(color.FgRed).SprintFunc()
		status := "Not defined"
		for _, value := range info.([]interface{}) {
			data := value.(map[string]interface{})
			if data["authUserInUserOwns"] == nil && data["authUserInRootOwns"] == nil {
				status = emoji.Sprint(":x:User - :x:Root")
			} else if data["authUserInUserOwns"] == true && data["authUserInRootOwns"] == nil {
				status = emoji.Sprint(":white_check_mark:User - :x:Root")
			} else if data["authUserInUserOwns"] == nil && data["authUserInRootOwns"] == true {
				status = emoji.Sprint(":x:User - :white_check_mark:Root")
			} else if data["authUserInUserOwns"] == true && data["authUserInRootOwns"] == true {
				status = emoji.Sprint(":white_check_mark:User - :white_check_mark:Root")
			}
			t, err := time.Parse(time.RFC3339Nano, data["release"].(string))
			if err != nil {
				fmt.Println("Erreur when date parsing :", err)
				return
			}
			datetime := t.Format("2006-01-02")
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", data["name"], data["os"], data["difficultyText"], data["user_owns_count"], data["root_owns_count"], data["stars"], status, datetime)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(activeCmd)
}
