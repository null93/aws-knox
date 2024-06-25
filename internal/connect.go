package internal

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/picker"
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
		now := time.Now()
		sessions, err := credentials.GetSessions()
		if err != nil {
			ExitWithError(1, "failed to get configured sessions", err)
		}
		if len(sessions) == 0 {
			ExitWithError(2, "no sso sessions found in config", err)
		}
		if connectSessionName == "" {
			p := picker.NewPicker()
			p.WithMaxHeight(10)
			p.WithEmptyMessage("No SSO Sessions Found")
			p.WithTitle("Pick SSO Session")
			p.WithHeaders("SSO Session", "Region", "SSO Start URL", "Expires In")
			for _, session := range sessions {
				expires := "-"
				if session.ClientToken != nil && !session.ClientToken.IsExpired() {
					expires = fmt.Sprintf("%.f mins", session.ClientToken.ExpiresAt.Sub(now).Minutes())
				}
				p.AddOption(session.Name, session.Name, session.Region, session.StartUrl, expires)
			}
			selection := p.Pick()
			if selection == nil {
				ExitWithError(3, "failed to pick an sso session", err)
			}
			connectSessionName = selection.Value.(string)
		}
		session := sessions.FindByName(connectSessionName)
		if session == nil {
			ExitWithError(4, "session with passed name not found", err)
		}
		if session.ClientToken == nil || session.ClientToken.IsExpired() {
			err := ClientLogin(session)
			if err != nil {
				ExitWithError(5, "failed to authorize device login", err)
			}
		}
		if connectAccountId == "" {
			connectAccountIds, err := session.GetAccounts()
			if err != nil {
				ExitWithError(6, "failed to get account ids", err)
			}
			if len(connectAccountIds) == 0 {
				ExitWithError(7, "no accounts found", err)
			}
			p := picker.NewPicker()
			p.WithMaxHeight(5)
			p.WithEmptyMessage("No Accounts Found")
			p.WithTitle("Pick Account")
			p.WithHeaders("Account ID", "Name", "Email")
			for _, account := range connectAccountIds {
				p.AddOption(account.Id, account.Id, account.Name, account.Email)
			}
			selection := p.Pick()
			if selection == nil {
				ExitWithError(8, "failed to pick an account id", err)
			}
			connectAccountId = selection.Value.(string)
		}
		roles, err := session.GetRoles(connectAccountId)
		if connectRoleName == "" {
			if err != nil {
				ExitWithError(9, "failed to get roles", err)
			}
			p := picker.NewPicker()
			p.WithMaxHeight(5)
			p.WithEmptyMessage("No Roles Found")
			p.WithTitle("Pick Role")
			p.WithHeaders("Role Name", "Expires In")
			for _, role := range roles {
				expires := "-"
				if role.Credentials != nil && !role.Credentials.IsExpired() {
					expires = fmt.Sprintf("%.f mins", role.Credentials.Expiration.Sub(now).Minutes())
				}
				p.AddOption(role.Name, role.Name, expires)
			}
			selection := p.Pick()
			if selection == nil {
				ExitWithError(10, "failed to pick a role name", err)
			}
			connectRoleName = selection.Value.(string)
		}
		role := roles.FindByName(connectRoleName)
		if role == nil {
			ExitWithError(11, "role with passed name not found", err)
		}
		if role.Credentials == nil || role.Credentials.IsExpired() {
			err := session.RefreshRoleCredentials(role)
			if err != nil {
				ExitWithError(12, "failed to get credentials", err)
			}
			err = role.Credentials.Save(session.Name, role.CacheKey())
			if err != nil {
				ExitWithError(13, "failed to save credentials", err)
			}
		}
		if err := role.MarkLastUsed(); err != nil {
			ExitWithError(14, "failed to mark last used role", err)
		}
		if connectInstanceId == "" {
			instances, err := role.GetManagedInstances()
			if err != nil {
				ExitWithError(15, "failed to get instances", err)
			}
			if len(instances) == 0 {
				ExitWithError(16, "no instances found", err)
			}
			p := picker.NewPicker()
			p.WithMaxHeight(10)
			p.WithEmptyMessage("No Instances Found")
			p.WithTitle("Pick EC2 Instance")
			p.WithHeaders("Instance ID", "Instance Type", "Private IP", "Public IP", "Name")
			for _, instance := range instances {
				p.AddOption(instance.Id, instance.Id, instance.InstanceType, instance.PrivateIpAddress, instance.PublicIpAddress, instance.Name)
			}
			selection := p.Pick()
			if selection == nil {
				ExitWithError(17, "failed to pick an instance id", err)
			}
			connectInstanceId = selection.Value.(string)
		}
		details, err := role.StartSession(connectInstanceId, connectUid)
		if err != nil {
			ExitWithError(18, "failed to start ssm session", err)
		}
		binaryPath, err := exec.LookPath("session-manager-plugin")
		if err != nil {
			ExitWithError(19, "failed to find session-manager-plugin, see "+SESSION_MANAGER_PLUGIN_URL, err)
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
		command.Stdin = cmd.InOrStdin()
		command.Stdout = cmd.OutOrStdout()
		command.Stderr = cmd.ErrOrStderr()
		err = command.Run()
		if err != nil {
			ExitWithError(20, "failed to run session-manager-plugin", err)
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
