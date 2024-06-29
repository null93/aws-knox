package internal

import (
	"fmt"
	"os"
	"syscall"

	"github.com/null93/aws-knox/pkg/ansi"
	"github.com/null93/aws-knox/pkg/color"
	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

var (
	Version     string = "0.0.0"
	debug       bool   = false
	view        string = "session"
	connectUid  uint32 = 0
	sessionName string
	accountId   string
	roleName    string
	instanceId  string
)

var RootCmd = &cobra.Command{
	Use:     "knox",
	Version: Version,
	Short:   "knox helps you manage AWS role credentials and connect to EC2 instances",
}

func ExitWithError(code int, message string, err error) {
	fmt.Printf("Error: %s\n", message)
	if err != nil && debug {
		fmt.Printf("Debug: %s\n", err.Error())
	}
	os.Exit(code)
}

func SelectRoleCredentialsStartingFromSession() (credentials.Sessions, *credentials.Session, credentials.Roles, *credentials.Role) {
	var err error
	var sessions credentials.Sessions
	var session *credentials.Session
	var roles credentials.Roles
	var role *credentials.Role
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
	return sessions, session, roles, role
}

func SelectRoleCredentialsStartingFromCache() (credentials.Sessions, *credentials.Session, *credentials.Role) {
	var err error
	var sessions credentials.Sessions
	var session *credentials.Session
	var role *credentials.Role
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
	return sessions, session, role
}

func init() {
	RootCmd.Flags().SortFlags = true
	RootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", debug, "Debug mode")

	if tty, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0); err == nil {
		ansi.Writer = tty
		color.Writer = tty
	}
}
