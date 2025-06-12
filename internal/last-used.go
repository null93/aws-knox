package internal

import (
	"fmt"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

var lastUsedCmd = &cobra.Command{
	Use:   "last-used",
	Short: "Use last used role credentials",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var sessions credentials.Sessions
		var session *credentials.Session
		var role credentials.Role
		var json string
		if role, err = credentials.GetLastUsedRole(); err != nil {
			ExitWithError(1, "failed to get last used role", err)
		}
		if role.Credentials == nil || role.Credentials.IsExpired() {
			if sessions, err = credentials.GetSessions(); err != nil {
				ExitWithError(2, "failed to parse sso sessions", err)
			}
			if session = sessions.FindByName(role.SessionName); session == nil {
				ExitWithError(3, "failed to find sso session "+role.SessionName, err)
			}
			if session.ClientToken == nil || session.ClientToken.IsExpired() {
				if err = tui.ClientLogin(session); err != nil {
					ExitWithError(4, "failed to authorize device login", err)
				}
			}
			if err = session.RefreshRoleCredentials(&role); err != nil {
				ExitWithError(5, "failed to get credentials", err)
			}
			if err = role.Credentials.Save(session.Name, role.CacheKey()); err != nil {
				ExitWithError(6, "failed to save credentials", err)
			}
		}
		if format == "env" {
			fmt.Printf("export AWS_ACCESS_KEY_ID=%q\n", role.Credentials.AccessKeyId)
			fmt.Printf("export AWS_SECRET_ACCESS_KEY=%q\n", role.Credentials.SecretAccessKey)
			fmt.Printf("export AWS_SESSION_TOKEN=%q\n", role.Credentials.SessionToken)
		} else {
			if json, err = role.Credentials.ToJSON(); err != nil {
				ExitWithError(7, "failed to serialize role credentials", err)
			} else {
				fmt.Println(json)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(lastUsedCmd)
	lastUsedCmd.Flags().SortFlags = true
	lastUsedCmd.Flags().StringVarP(&format, "format", "f", format, "Output format (json or env)")
}
