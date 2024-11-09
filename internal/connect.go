package internal

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/null93/aws-knox/pkg/color"
	"github.com/null93/aws-knox/sdk/credentials"
	. "github.com/null93/aws-knox/sdk/style"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <instance-search-term>",
	Short: "Connect to an EC2 instance using session-manager-plugin",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchTerm := ""
		currentSelector := "instance"
		if len(args) > 0 {
			searchTerm = args[0]
		}
		var err error
		var role *credentials.Role
		var action string
		var binaryPath string
		if lastUsed {
			var err error
			var sessions credentials.Sessions
			var session *credentials.Session
			var roleTemp credentials.Role
			if roleTemp, err = credentials.GetLastUsedRole(); err != nil {
				ExitWithError(1, "failed to get last used role", err)
			}
			role = &roleTemp
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
				if err = session.RefreshRoleCredentials(role); err != nil {
					ExitWithError(5, "failed to get credentials", err)
				}
				if err = role.Credentials.Save(session.Name, role.CacheKey()); err != nil {
					ExitWithError(6, "failed to save credentials", err)
				}
			}
		}
		for {
			if role == nil {
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
					}
					continue
				}
			}
			if region == "" {
				region = role.Region
			}
			if instanceId == "" {
				if currentSelector == "instance" {
					if instanceId, action, err = tui.SelectInstance(role, region, searchTerm, instanceColTags); err != nil {
						ExitWithError(19, "failed to pick an instance", err)
					} else if action == "back" {
						goBack(&role)
						continue
					} else if action == "pick-region" {
						currentSelector = "region"
						continue
					}
				} else {
					pickedRegion := ""
					if pickedRegion, action, err = tui.SelectRegion(region); err != nil {
						ExitWithError(20, "failed to pick a region", err)
					} else if action == "back" {
						currentSelector = "instance"
						continue
					}
					currentSelector = "instance"
					region = pickedRegion
					continue
				}
			}

			yellow := color.ToForeground(YellowColor).Decorator()
			gray := color.ToForeground(LightGrayColor).Decorator()
			title := TitleStyle.Decorator()
			DefaultStyle.Printfln("")
			DefaultStyle.Printfln("%s %s", title("SSO Session: "), gray(role.SessionName))
			DefaultStyle.Printfln("%s %s", title("Region:      "), gray(region))
			DefaultStyle.Printfln("%s %s", title("Account ID:  "), gray(role.AccountId))
			DefaultStyle.Printfln("%s %s", title("Role Name:   "), gray(role.Name))
			DefaultStyle.Printfln("%s %s", title("Instance ID: "), yellow(instanceId))

			details, err := role.StartSession(instanceId, connectUid)
			if err != nil {
				ExitWithError(20, "failed to start ssm session", err)
			}
			if binaryPath, err = exec.LookPath("session-manager-plugin"); err != nil {
				ExitWithError(21, "failed to find session-manager-plugin, see "+SESSION_MANAGER_PLUGIN_URL, err)
			}
			command := exec.Command(
				binaryPath,
				fmt.Sprintf(`{"SessionId": "%s", "TokenValue": "%s", "StreamUrl": "%s"}`, *details.SessionId, *details.TokenValue, *details.StreamUrl),
				region,
				"StartSession",
				"", // No Profile
				fmt.Sprintf(`{"Target": "%s"}`, instanceId),
				fmt.Sprintf("https://ssm.%s.amazonaws.com", region),
			)
			command.Stdin = os.Stdin
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
			if err = command.Run(); err != nil {
				ExitWithError(22, "failed to run session-manager-plugin", err)
			}
			break
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
	connectCmd.Flags().StringVar(&region, "region", region, "Region for quering instances")
	connectCmd.Flags().BoolVarP(&lastUsed, "last-used", "l", lastUsed, "select last used credentials")
	connectCmd.Flags().Uint32VarP(&connectUid, "uid", "u", connectUid, "UID on instance to 'su' to")
}
