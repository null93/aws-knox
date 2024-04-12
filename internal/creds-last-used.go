package internal

import (
	"fmt"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/spf13/cobra"
)

var credsLastUsedCmd = &cobra.Command{
	Use:   "last-used",
	Short: "Use last used role credentials",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		role, err := credentials.GetLastUsedRole()
		if err != nil {
			ExitWithError(1, "failed to get last used role", err)
		}
		serialized, err := role.Credentials.ToJSON()
		if err != nil {
			ExitWithError(2, "failed to serialize role credentials", err)
		}
		fmt.Println(serialized)
	},
}

func init() {
	credsCmd.AddCommand(credsLastUsedCmd)
	credsLastUsedCmd.Flags().SortFlags = true
}
