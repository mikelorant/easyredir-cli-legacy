package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mikelorant/easyredir-cli/internal/importer"
)

var (
	importFile    string
	importFormat  string
	importPreview bool

	importCmd = &cobra.Command{
		Use:   "import",
		Short: "A brief description of your command",
	}

	importRulesCmd = &cobra.Command{
		Use:   "rules",
		Short: "A brief description of your command",
		Run: func(cmd *cobra.Command, args []string) {
			doImportRules()
		},
	}
)

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.AddCommand(importRulesCmd)
	importRulesCmd.Flags().BoolVarP(&importPreview, "preview", "", false, "Preview")
	importRulesCmd.Flags().StringVarP(&importFile, "file", "", "", "Filename")
	importRulesCmd.Flags().StringVarP(&importFormat, "format", "", "", "Filename")
	importRulesCmd.MarkFlagRequired("file")
}

func doImportRules() {
	importer.Import(&importer.Options{
		File:    importFile,
		Format:  importFormat,
		Preview: importPreview,
	})
}
