package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/null93/aws-knox/pkg/color"
	"github.com/null93/aws-knox/sdk/credentials"
	. "github.com/null93/aws-knox/sdk/style"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

const RSYNC_CONFIG = `
uid = 0
gid = 0
use chroot = yes
read only = false
hosts allow = 127.0.0.1

[sync]
path = /root/knox-sync
comment = knox sync
`

const RSYNC_INIT_SCRIPT = `
if ! command -v rsync > /dev/null; then echo "EXIT_CODE: 45"; exit 45; fi;
if ! [ -d /root/knox-sync ]; then mkdir -p /root/knox-sync; fi;
if [ -f /run/knox-rsyncd.pid ]; then echo "EXIT_CODE: 46"; exit 46; fi;
`

var (
	rsyncPort uint16 = 9999
	localPort uint16 = 8080
)

func rsyncInit(role *credentials.Role, instanceId string) {
	var binaryPath string
	details, err := role.StartCommand(instanceId, strings.ReplaceAll(RSYNC_INIT_SCRIPT, "\n", " "))
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
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
	if output, err := command.CombinedOutput(); err != nil {
		ExitWithError(22, "failed to run session-manager-plugin", err)
	} else if strings.Contains(string(output), "EXIT_CODE: 45") {
		ExitWithError(45, "rsync is not installed on the target instance", nil)
	} else if strings.Contains(string(output), "EXIT_CODE: 46") {
		fmt.Println("OUTPUT", string(output))
		ExitWithError(46, "rsyncd is already running on the target instance", nil)
	}
	if debug {
		fmt.Println("Debug: rsync detected on the target instance")
		fmt.Println("Debug: ensuring /root/knox-sync folder exists")
		fmt.Println("Debug: making sure another rsync daemon is not running")
	}
}

func rsyncStart(role *credentials.Role, instanceId string) {
	var binaryPath string
	var configEncoded = base64.StdEncoding.EncodeToString([]byte(RSYNC_CONFIG))
	var startCommand = fmt.Sprintf("rsync --daemon --port=%d --log-file=/dev/null --config=/tmp/knox-rsyncd.conf --dparam=pidfile=/run/knox-rsyncd.pid", rsyncPort)
	details, err := role.StartCommand(instanceId, fmt.Sprintf("echo %s | base64 -d > /tmp/knox-rsyncd.conf; %s || (echo 'EXIT_CODE: 47' || exit 47)", configEncoded, startCommand))
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
	command.Stdout = nil
	command.Stderr = nil
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
	if err = command.Run(); err != nil {
		ExitWithError(22, "failed to run session-manager-plugin", err)
	}
	if debug {
		fmt.Printf("Debug: rsync daemon started on port %d\n", rsyncPort)
	}
}

func rsyncClean(role *credentials.Role, instanceId string) {
	var binaryPath string
	removeOld := "rm -f /tmp/knox-rsyncd.conf;"
	killOld := "if [ -f /run/knox-rsyncd.pid ]; then kill -9 $(cat /run/knox-rsyncd.pid); rm -f /run/knox-rsyncd.pid; fi;"
	details, err := role.StartCommand(instanceId, removeOld+killOld)
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
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
	if _, err := command.CombinedOutput(); err != nil {
		ExitWithError(22, "failed to run session-manager-plugin", err)
	}
	if debug {
		fmt.Println("Debug: removing old temporary rsync config")
		fmt.Println("Debug: making sure old rsync daemon is not running")
	}
}

func rsyncPortForward(role *credentials.Role, instanceId string) {
	var binaryPath string
	details, err := role.StartPortForward(instanceId, rsyncPort, localPort)
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
	command.Stdout = nil
	command.Stderr = nil
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
	if err = command.Run(); err != nil {
		ExitWithError(22, "failed to run session-manager-plugin", err)
	}
}

var syncCmd = &cobra.Command{
	Use:   "sync [instance-search-term]",
	Short: "start rsyncd and port forward to it",
	Run: func(cmd *cobra.Command, args []string) {
		searchTerm := strings.Join(args, " ")
		currentSelector := "instance"
		var err error
		var role *credentials.Role
		var action string
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
						role = nil
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
					} else if action == "refresh" {
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
			DefaultStyle.Printfln("%s %s", title("SSO Session:        "), gray(role.SessionName))
			DefaultStyle.Printfln("%s %s", title("Region:             "), gray(region))
			DefaultStyle.Printfln("%s %s", title("Account ID:         "), gray(role.AccountId))
			DefaultStyle.Printfln("%s %s", title("Role Name:          "), gray(role.Name))
			DefaultStyle.Printfln("%s %s", title("Instance ID:        "), gray(instanceId))
			DefaultStyle.Printfln("%s %s", title("Remote Destination: "), gray("/root/knox-sync"))
			DefaultStyle.Printfln("%s %s", title("Example Command:    "), yellow("rsync -P ./dump.sql ./release.tar.gz rsync://127.0.0.1:%d/sync\n", localPort))

			defer rsyncClean(role, instanceId)
			defer func() {
				fmt.Println("\nCleaning up...")
			}()
			fmt.Println("Starting rsync daemon on target instance...")
			rsyncClean(role, instanceId)
			rsyncInit(role, instanceId)
			rsyncStart(role, instanceId)
			fmt.Printf("Port forwarding to 127.0.0.1:%d...\n", localPort)
			rsyncPortForward(role, instanceId)
			break
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)
	syncCmd.Flags().SortFlags = true
	syncCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	syncCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	syncCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
	syncCmd.Flags().StringVarP(&instanceId, "instance-id", "i", instanceId, "EC2 instance ID")
	syncCmd.Flags().StringVar(&region, "region", region, "Region for quering instances")
	syncCmd.Flags().Uint16VarP(&rsyncPort, "rsync-port", "P", rsyncPort, "rsync port")
	syncCmd.Flags().Uint16VarP(&localPort, "local-port", "p", localPort, "local port")
	syncCmd.Flags().BoolVarP(&lastUsed, "last-used", "l", lastUsed, "select last used credentials")
}
