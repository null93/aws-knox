package internal

import (
	"fmt"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
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
		var lastRole credentials.Role
		var action string
		var sessions credentials.Sessions
		var session *credentials.Session
		var err error

		for {
			if lastUsed {
				if lastRole, err = credentials.GetLastUsedRole(); err != nil {
					ExitWithError(1, "failed to get last used role", err)
				}
				if lastRole.Credentials == nil || lastRole.Credentials.IsExpired() {
					if sessions, err = credentials.GetSessions(); err != nil {
						ExitWithError(2, "failed to parse sso sessions", err)
					}
					if session = sessions.FindByName(lastRole.SessionName); session == nil {
						ExitWithError(3, "failed to find sso session "+lastRole.SessionName, err)
					}
					if session.ClientToken == nil || session.ClientToken.IsExpired() {
						if err = tui.ClientLogin(session); err != nil {
							ExitWithError(4, "failed to authorize device login", err)
						}
					}
					if err = session.RefreshRoleCredentials(&lastRole); err != nil {
						ExitWithError(5, "failed to get credentials", err)
					}
					if !doNotCache {
						if err = lastRole.Credentials.Save(session.Name, lastRole.CacheKey()); err != nil {
							ExitWithError(6, "failed to save credentials", err)
						}
					}
				}
			} else {
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
			}
			targetRole := role
			if lastUsed {
				targetRole = &lastRole
			}
			if format == "env" {
				fmt.Printf("export AWS_ACCESS_KEY_ID=%q\n", targetRole.Credentials.AccessKeyId)
				fmt.Printf("export AWS_SECRET_ACCESS_KEY=%q\n", targetRole.Credentials.SecretAccessKey)
				fmt.Printf("export AWS_SESSION_TOKEN=%q\n", targetRole.Credentials.SessionToken)
				break
			} else {
				if json, err := targetRole.Credentials.ToJSON(); err != nil {
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
	selectCmd.Flags().BoolVarP(&lastUsed, "last-used", "l", lastUsed, "Use last used role credentials")
}
