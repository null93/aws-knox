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
	Run: func(cmd *cobra.Command, args []string) {
		var role *credentials.Role
		var action string
		for {
			if !selectCachedFirst || (sessionName != "" && accountId != "" && roleName != "") {
				action, role = SelectRoleCredentialsStartingFromSession()
			} else {
				action, role = SelectRoleCredentialsStartingFromCache()
			}
			if action == "toggle-view" {
				toggleView()
				continue
			}
			if action == "back" {
				goBack(&role)
				continue
			}
			if action == "delete" {
				if role != nil && role.Credentials != nil {
					role.Credentials.DeleteCache(role.SessionName, role.CacheKey())
					role = nil
				}
				continue
			}
			if json, err := role.Credentials.ToJSON(); err != nil {
				ExitWithError(19, "failed to convert credentials to json", err)
			} else {
				fmt.Println(json)
				break
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)
	selectCmd.Flags().SortFlags = true
	selectCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	selectCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	selectCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
}
