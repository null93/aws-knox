package internal

import (
	"fmt"
	"time"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/picker"
	"github.com/spf13/cobra"
)

var credsSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Pick from cached role credentials",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		roles, err := credentials.GetSavedRolesWithCredentials()
		if err != nil {
			ExitWithError(1, "failed to get role credentials", err)
		}
		p := picker.NewPicker()
		p.WithMaxHeight(10)
		p.WithEmptyMessage("No Role Credentials Found")
		p.WithTitle("Pick Role Credentials")
		p.WithHeaders("Region", "Account ID", "Role Name", "Expires In")
		for _, role := range roles {
			expires := "-"
			if role.Credentials != nil && !role.Credentials.IsExpired() {
				expires = fmt.Sprintf("%.f mins", role.Credentials.Expiration.Sub(now).Minutes())
			}
			p.AddOption(role, role.Region, role.AccountId, role.Name, expires)
		}
		selection := p.Pick()
		if selection == nil {
			ExitWithError(3, "failed to pick role credentials", err)
		}
		selectedRole := selection.Value.(credentials.Role)
		serialized, err := selectedRole.Credentials.ToJSON()
		if err != nil {
			ExitWithError(4, "failed to serialize role credentials", err)
		}
		if err := selectedRole.MarkLastUsed(); err != nil {
			ExitWithError(5, "failed to mark last used role", err)
		}
		fmt.Println(serialized)
	},
}

func init() {
	credsCmd.AddCommand(credsSelectCmd)
	credsSelectCmd.Flags().SortFlags = true
}
