package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/tui"
	"github.com/spf13/cobra"
)

const RSYNC_CONFIG = `
uid = 0
gid = 0
use chroot = yes
read only = false

[upload]
path = /root/upload
comment = knox upload
`

const RSYNC_INIT_SCRIPT = `
if ! command -v rsync > /dev/null; then echo "EXIT_CODE: 45"; exit 45; fi;
if ! [ -d /root/upload ]; then mkdir -p /root/upload; fi;
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
		role.Region,
		"StartSession",
		"", // No Profile
		fmt.Sprintf(`{"Target": "%s"}`, instanceId),
		fmt.Sprintf("https://ssm.%s.amazonaws.com", role.Region),
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
		fmt.Println("Debug: ensuring /root/upload folder exists")
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
		role.Region,
		"StartSession",
		"", // No Profile
		fmt.Sprintf(`{"Target": "%s"}`, instanceId),
		fmt.Sprintf("https://ssm.%s.amazonaws.com", role.Region),
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
		role.Region,
		"StartSession",
		"", // No Profile
		fmt.Sprintf(`{"Target": "%s"}`, instanceId),
		fmt.Sprintf("https://ssm.%s.amazonaws.com", role.Region),
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
		role.Region,
		"StartSession",
		"", // No Profile
		fmt.Sprintf(`{"Target": "%s"}`, instanceId),
		fmt.Sprintf("https://ssm.%s.amazonaws.com", role.Region),
	)
	command.Stdin = os.Stdin
	command.Stdout = nil
	command.Stderr = nil
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Foreground: true}
	if err = command.Run(); err != nil {
		ExitWithError(22, "failed to run session-manager-plugin", err)
	}
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "start rsyncd and port forward to it",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var role *credentials.Role
		var action string
		for {
			if !selectCachedFirst {
				action, role = SelectRoleCredentialsStartingFromSession()
			} else {
				action, role = SelectRoleCredentialsStartingFromCache()
			}
			if action == "toggle-view" {
				toggleView()
				continue
			}
			if action == "back" {
				goBack()
				continue
			}
			if action == "delete" {
				if role != nil && role.Credentials != nil {
					role.Credentials.DeleteCache(role.SessionName, role.CacheKey())
				}
				continue
			}
			if instanceId == "" {
				if instanceId, action, err = tui.SelectInstance(role); err != nil {
					ExitWithError(19, "failed to pick an instance", err)
				} else if action == "back" {
					goBack()
					continue
				}
			}
			fmt.Println("Remote Destination:  /root/upload")
			fmt.Printf("Example Command:     rsync -P ./dump.sql ./release.tar.gz rsync://127.0.0.1:%d/upload\n", localPort)
			fmt.Println()
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
	RootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().SortFlags = true
	uploadCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	uploadCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	uploadCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
	uploadCmd.Flags().StringVarP(&instanceId, "instance-id", "i", instanceId, "EC2 instance ID")
	uploadCmd.Flags().Uint16VarP(&rsyncPort, "rsync-port", "P", rsyncPort, "rsync port")
	uploadCmd.Flags().Uint16VarP(&localPort, "local-port", "p", localPort, "local port")
	uploadCmd.Flags().BoolVarP(&selectCachedFirst, "cached", "c", selectCachedFirst, "select from cached credentials")
}
