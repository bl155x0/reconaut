package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "reconaut -t <template-action> VARIABLE=VALUE [VARIABLE=VALUE...]",
	Short: `
 ██▀███   ▓█████  ▄████▄  ▒█████   ███▄    █  ▄▄▄      █    ██ ▄▄▄█████▓ 💀
▓██ ▒ ██▒ ▓█   ▀ ▒██▀ ▀█ ▒██▒  ██▒ ██ ▀█   █ ▒████▄    ██  ▓██▒▓  ██▒ ▓▒ 
▓██ ░▄█ ▒ ▒███   ▒▓█    ▄▒██░  ██▒▓██  ▀█ ██▒▒██  ▀█▄ ▓██  ▒██░▒ ▓██░ ▒░
▒██▀▀█▄   ▒▓█  ▄▒▒▓▓▄ ▄██▒██   ██░▓██▒  ▐▌██▒░██▄▄▄▄██▓▓█  ░██░░ ▓██▓ ░ 
░██▓ ▒██▒▒░▒████░▒ ▓███▀ ░ ████▓▒░▒██░   ▓██░▒▓█   ▓██▒▒█████▓   ▒██▒ ░ 
░ ▒▓ ░▒▓░░░░ ▒░ ░░ ░▒ ▒  ░ ▒░▒░▒░ ░ ▒░   ▒ ▒ ░▒▒   ▓▒█░▒▓▒ ▒ ▒   ▒ ░░   
  ░▒ ░ ▒░░ ░ ░     ░  ▒    ░ ▒ ▒░ ░ ░░   ░ ▒░░ ░   ▒▒ ░░▒░ ░ ░     ░    
   ░   ░     ░   ░       ░ ░ ░ ▒     ░   ░ ░   ░   ▒   ░░░ ░ ░   ░      
   ░     ░   ░   ░ ░         ░ ░           ░       ░     ░              

    	       basic reconnaissance job processor
	                    ★ bl155 ★
`,
	Run: func(cmd *cobra.Command, args []string) {
		template, err := cmd.Flags().GetString("template")
		cobra.CheckErr(err)

		projectFile, err := cmd.Flags().GetString("project-name")
		cobra.CheckErr(err)

		verbose, err := cmd.Flags().GetBool("verbose")
		cobra.CheckErr(err)

		err = start(projectFile, template, verbose, args)
		cobra.CheckErr(err)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("project-name", "p", "", "The project base name (database, etc.)")
	rootCmd.Flags().StringP("template", "t", "", "The action to to execute (the template to run) (mandatory)")
	rootCmd.Flags().BoolP("verbose", "v", false, "Be more verbose in the output")
	rootCmd.MarkFlagRequired("template")
}
