package internal

import (
	"fmt"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select specific AWS role credentials",
	Args:  cobra.ExactArgs(0),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if view != "session" && view != "cached" {
			return fmt.Errorf("view must be either 'session' or 'cached'")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var role *credentials.Role
		if view == "session" {
			_, _, _, role = SelectRoleCredentialsStartingFromSession()
		} else {
			_, _, role = SelectRoleCredentialsStartingFromCache()
		}
		if json, err := role.Credentials.ToJSON(); err != nil {
			ExitWithError(12, "failed to convert credentials to json", err)
		} else {
			fmt.Println(json)
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)
	selectCmd.Flags().SortFlags = true
	selectCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	selectCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	selectCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
	selectCmd.Flags().StringVarP(&view, "view", "v", view, "session or cached")
}
