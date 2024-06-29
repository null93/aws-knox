package internal

import (
	"fmt"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

var (
	sessionName string
	accountId   string
	roleName    string
	view        string = "session"
)

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select specific AWS role credentials starting from SSO",
	Args:  cobra.ExactArgs(0),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if view != "session" && view != "cached" {
			return fmt.Errorf("view must be either 'session' or 'cached'")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var sessions credentials.Sessions
		var session *credentials.Session
		var roles credentials.Roles
		var role *credentials.Role
		var json string
		if sessions, err = credentials.GetSessions(); err != nil {
			ExitWithError(1, "failed to get configured sessions", err)
		}
		if sessionName == "" {
			if sessionName, err = tui.SelectSession(sessions); err != nil {
				ExitWithError(2, "failed to pick an sso session", err)
			}
		}
		if session = sessions.FindByName(sessionName); session == nil {
			ExitWithError(3, "session with passed name not found", err)
		}
		if session.ClientToken == nil || session.ClientToken.IsExpired() {
			if err = tui.ClientLogin(session); err != nil {
				ExitWithError(4, "failed to authorize device login", err)
			}
		}
		if accountId == "" {
			if accountId, err = tui.SelectAccount(session); err != nil {
				ExitWithError(5, "failed to pick an account id", err)
			}
		}
		if roles, err = session.GetRoles(accountId); err != nil {
			ExitWithError(6, "failed to get roles", err)
		}
		if roleName == "" {
			if roleName, err = tui.SelectRole(roles); err != nil {
				ExitWithError(7, "failed to pick a role", err)
			}
		}
		if role = roles.FindByName(roleName); role == nil {
			ExitWithError(8, "role with passed name not found", err)
		}
		if role.Credentials == nil || role.Credentials.IsExpired() {
			if err = session.RefreshRoleCredentials(role); err != nil {
				ExitWithError(9, "failed to get credentials", err)
			}
			if err = role.Credentials.Save(session.Name, role.CacheKey()); err != nil {
				ExitWithError(10, "failed to save credentials", err)
			}
		}
		if err := role.MarkLastUsed(); err != nil {
			ExitWithError(11, "failed to mark last used role", err)
		}
		if json, err = role.Credentials.ToJSON(); err != nil {
			ExitWithError(12, "failed to convert credentials to json", err)
		}
		fmt.Println(json)
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
