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
		if format != "json" && format != "env" {
			return fmt.Errorf("invalid format: %s, must be 'json' or 'env'", format)
		}
		return nil
	},
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
			if format == "env" {
				fmt.Printf("export AWS_ACCESS_KEY_ID=%q\n", role.Credentials.AccessKeyId)
				fmt.Printf("export AWS_SECRET_ACCESS_KEY=%q\n", role.Credentials.SecretAccessKey)
				fmt.Printf("export AWS_SESSION_TOKEN=%q\n", role.Credentials.SessionToken)
				break
			} else {
				if json, err := role.Credentials.ToJSON(); err != nil {
					ExitWithError(19, "failed to convert credentials to json", err)
				} else {
					fmt.Println(json)
					break
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)
	selectCmd.Flags().SortFlags = true
	selectCmd.Flags().StringVarP(&format, "format", "f", format, "Output format (json or env)")
	selectCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	selectCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	selectCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
	selectCmd.Flags().BoolVarP(&doNotCache, "no-cache", "n", doNotCache, "Do not cache credentials")
}
