package cmd

import (
	"fmt"
	"reconaut/framework/job"
)

func listTemplates() error {
	fmt.Printf("\n")
	for _, file := range job.GetTemplateFiles() {
		template, err := job.NewTemplateFromFile(file)
		if err != nil {
			return err
		}

		fmt.Printf(" - %s [%s]\n", template.Name, file)
		fmt.Printf("   %s", template.Description)
		fmt.Printf("\n")
	}
	return nil
}
