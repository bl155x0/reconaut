package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	listMode bool
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
	PreRun: func(cmd *cobra.Command, _ []string) {
		if listMode == false {
			cmd.MarkFlagRequired("template")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		template, err := cmd.Flags().GetString("template")
		cobra.CheckErr(err)

		projectFile, err := cmd.Flags().GetString("project-name")
		cobra.CheckErr(err)

		verbose, err := cmd.Flags().GetBool("verbose")
		cobra.CheckErr(err)

		list, err := cmd.Flags().GetBool("list")
		cobra.CheckErr(err)

		if list {
			//list only the available templates
			listTemplates()
		} else {
			//start the recon
			err = start(projectFile, template, verbose, args)
			cobra.CheckErr(err)
		}
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
	rootCmd.Flags().BoolVarP(&listMode, "list", "l", false, "Lists the available templates")
	rootCmd.Flags().BoolP("verbose", "v", false, "Be more verbose in the output")
}
