package internal

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

const (
	SESSION_MANAGER_PLUGIN_URL = "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"
)

var (
	connectSessionName string
	connectAccountId   string
	connectRoleName    string
	connectInstanceId  string
	connectUid         uint32 = 0
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a specific EC2 instance using AWS session-manager-plugin",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var sessions credentials.Sessions
		var session *credentials.Session
		var roles credentials.Roles
		var role *credentials.Role
		var binaryPath string
		if sessions, err = credentials.GetSessions(); err != nil {
			ExitWithError(1, "failed to get configured sessions", err)
		}
		if connectSessionName == "" {
			if connectSessionName, err = tui.SelectSession(sessions); err != nil {
				ExitWithError(2, "failed to pick an sso session", err)
			}
		}
		if session = sessions.FindByName(connectSessionName); session == nil {
			ExitWithError(3, "session with passed name not found", err)
		}
		if session.ClientToken == nil || session.ClientToken.IsExpired() {
			if err = tui.ClientLogin(session); err != nil {
				ExitWithError(4, "failed to authorize device login", err)
			}
		}
		if connectAccountId == "" {
			if connectAccountId, err = tui.SelectAccount(session); err != nil {
				ExitWithError(5, "failed to pick an account id", err)
			}
		}
		if roles, err = session.GetRoles(connectAccountId); err != nil {
			ExitWithError(6, "failed to get roles", err)
		}
		if connectRoleName == "" {
			if connectRoleName, err = tui.SelectRole(roles); err != nil {
				ExitWithError(7, "failed to pick a role", err)
			}
		}
		if role = roles.FindByName(connectRoleName); role == nil {
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
		if connectInstanceId == "" {
			if connectInstanceId, err = tui.SelectInstance(role); err != nil {
				ExitWithError(12, "failed to pick an instance", err)
			}
		}
		details, err := role.StartSession(connectInstanceId, connectUid)
		if err != nil {
			ExitWithError(13, "failed to start ssm session", err)
		}
		if binaryPath, err = exec.LookPath("session-manager-plugin"); err != nil {
			ExitWithError(14, "failed to find session-manager-plugin, see "+SESSION_MANAGER_PLUGIN_URL, err)
		}
		command := exec.Command(
			binaryPath,
			fmt.Sprintf(`{"SessionId": "%s", "TokenValue": "%s", "StreamUrl": "%s"}`, *details.SessionId, *details.TokenValue, *details.StreamUrl),
			role.Region,
			"StartSession",
			"", // No Profile
			fmt.Sprintf(`{"Target": "%s"}`, connectInstanceId),
			fmt.Sprintf("https://ssm.%s.amazonaws.com", role.Region),
		)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
		if err = command.Run(); err != nil {
			ExitWithError(15, "failed to run session-manager-plugin", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(connectCmd)
	connectCmd.Flags().SortFlags = true
	connectCmd.Flags().StringVarP(&connectSessionName, "sso-session", "s", connectSessionName, "SSO session name")
	connectCmd.Flags().StringVarP(&connectAccountId, "account-id", "a", connectAccountId, "AWS account ID")
	connectCmd.Flags().StringVarP(&connectRoleName, "role-name", "r", connectRoleName, "AWS role name")
	connectCmd.Flags().StringVarP(&connectInstanceId, "instance-id", "i", connectInstanceId, "EC2 instance ID")
	connectCmd.Flags().Uint32VarP(&connectUid, "uid", "u", connectUid, "UID on instance to 'su' to")
}
