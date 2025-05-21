package cmd

import (
	"fmt"
	"github.com/defektive/frenzy/pkg/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
)

// configCmd represents the base command when called without any subcommands
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "show config",
	Long:  `show config`,
	Run: func(cmd *cobra.Command, args []string) {
		yamlBytes, err := yaml.Marshal(server.GetConfig())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(yamlBytes))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
