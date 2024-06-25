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
		if role.Credentials == nil || role.Credentials.IsExpired() {
			sessions, err := credentials.GetSessions()
			if err != nil {
				ExitWithError(2, "failed to parse sso sessions", err)
			}
			session := sessions.FindByName(role.SessionName)
			if session == nil {
				ExitWithError(3, "failed to find sso session "+role.SessionName, err)
			}
			if session.ClientToken == nil || session.ClientToken.IsExpired() {
				err := ClientLogin(session)
				if err != nil {
					ExitWithError(4, "failed to authorize device login", err)
				}
			}
			err = session.RefreshRoleCredentials(&role)
			if err != nil {
				ExitWithError(5, "failed to get credentials", err)
			}
			err = role.Credentials.Save(session.Name, role.CacheKey())
			if err != nil {
				ExitWithError(6, "failed to save credentials", err)
			}
		}
		serialized, err := role.Credentials.ToJSON()
		if err != nil {
			ExitWithError(7, "failed to serialize role credentials", err)
		}
		fmt.Println(serialized)
	},
}

func init() {
	credsCmd.AddCommand(credsLastUsedCmd)
	credsLastUsedCmd.Flags().SortFlags = true
}
