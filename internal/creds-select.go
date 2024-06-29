package internal

import (
	"fmt"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

var credsSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Pick from cached role credentials",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var sessions credentials.Sessions
		var session *credentials.Session
		var role *credentials.Role
		var json string
		role, err = tui.SelectRolesCredentials()
		if role.Credentials == nil || role.Credentials.IsExpired() {
			if sessions, err = credentials.GetSessions(); err != nil {
				ExitWithError(1, "failed to parse sso sessions", err)
			}
			if session = sessions.FindByName(role.SessionName); session == nil {
				ExitWithError(2, "failed to find sso session "+role.SessionName, err)
			}
			if session.ClientToken == nil || session.ClientToken.IsExpired() {
				if err = tui.ClientLogin(session); err != nil {
					ExitWithError(3, "failed to authorize device login", err)
				}
			}

			if err = session.RefreshRoleCredentials(role); err != nil {
				ExitWithError(9, "failed to get credentials", err)
			}
			if err = role.Credentials.Save(session.Name, role.CacheKey()); err != nil {
				ExitWithError(10, "failed to save credentials", err)
			}
		}
		if err = role.MarkLastUsed(); err != nil {
			ExitWithError(11, "failed to mark last used role", err)
		}
		if json, err = role.Credentials.ToJSON(); err != nil {
			ExitWithError(12, "failed to serialize role credentials", err)
		}
		fmt.Println(json)
	},
}

func init() {
	credsCmd.AddCommand(credsSelectCmd)
	credsSelectCmd.Flags().SortFlags = true
}
