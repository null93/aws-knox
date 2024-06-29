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

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an EC2 instance using session-manager-plugin",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var role *credentials.Role
		var binaryPath string
		if view == "session" {
			_, _, _, role = SelectRoleCredentialsStartingFromSession()
		} else {
			_, _, role = SelectRoleCredentialsStartingFromCache()
		}
		if instanceId == "" {
			if instanceId, err = tui.SelectInstance(role); err != nil {
				ExitWithError(12, "failed to pick an instance", err)
			}
		}
		details, err := role.StartSession(instanceId, connectUid)
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
			fmt.Sprintf(`{"Target": "%s"}`, instanceId),
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
	connectCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	connectCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	connectCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
	connectCmd.Flags().StringVarP(&instanceId, "instance-id", "i", instanceId, "EC2 instance ID")
	connectCmd.Flags().StringVarP(&view, "view", "v", view, "session or cached")
	connectCmd.Flags().Uint32VarP(&connectUid, "uid", "u", connectUid, "UID on instance to 'su' to")
}
