package internal

import (
	"github.com/spf13/cobra"
)

var credsCmd = &cobra.Command{
	Use:   "creds",
	Short: "Interact with generated role credentials",
}

func init() {
	RootCmd.AddCommand(credsCmd)
	credsCmd.Flags().SortFlags = true
}
