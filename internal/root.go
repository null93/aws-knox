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
	"github.com/spf13/viper"
)

var (
	Version           string = "0.0.0"
	debug             bool   = false
	selectCachedFirst bool   = false
	connectUid        uint32 = 0
	lastUsed          bool   = false
	sessionName       string
	accountId         string
	roleName          string
	instanceId        string
	region            string
	accountAliases    map[string]string
	instanceColTags   []string
)

const (
	SESSION_MANAGER_PLUGIN_URL = "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"
)

var RootCmd = &cobra.Command{
	Use:     "knox",
	Version: Version,
	Short:   "knox helps you manage AWS role credentials and connect to EC2 instances",
}

func toggleView() {
	selectCachedFirst = !selectCachedFirst
}

func goBack(role **credentials.Role) {
	if lastUsed {
		lastUsed = false
	}
	if *role != nil {
		*role = nil
		roleName = ""
		return
	}
	if instanceId != "" {
		instanceId = ""
		return
	}
	if instanceId != "" {
		instanceId = ""
		return
	}
	if roleName != "" {
		roleName = ""
		return
	}
	if accountId != "" {
		accountId = ""
		return
	}
	if sessionName != "" {
		sessionName = ""
	}
}

func ExitWithError(code int, message string, err error) {
	fmt.Printf("Error: %s\n", message)
	if err != nil && debug {
		fmt.Printf("Debug: %s\n", err.Error())
	} else {
		fmt.Printf("Info:  run with --debug flag for more info\n")
	}
	os.Exit(code)
}

func SelectRoleCredentialsStartingFromSession() (string, *credentials.Role) {
	var err error
	var action string
	var sessions credentials.Sessions
	var session *credentials.Session
	var roles credentials.Roles
	var role *credentials.Role
	if sessions, err = credentials.GetSessions(); err != nil {
		ExitWithError(1, "failed to get configured sessions", err)
	}
	if sessionName == "" {
		if sessionName, action, err = tui.SelectSession(sessions); err != nil {
			ExitWithError(2, "failed to pick an sso session", err)
		} else if action != "" {
			return action, nil
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
		if accountId, action, err = tui.SelectAccount(session, accountAliases); err != nil {
			ExitWithError(5, "failed to pick an account id", err)
		} else if action != "" {
			return action, nil
		}
	}
	if roles, err = session.GetRoles(accountId); err != nil {
		ExitWithError(6, "failed to get roles", err)
	}
	if roleName == "" {
		if roleName, action, err = tui.SelectRole(roles); err != nil {
			ExitWithError(7, "failed to pick a role", err)
		} else if action != "" {
			return action, nil
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
	return "", role
}

func SelectRoleCredentialsStartingFromCache() (string, *credentials.Role) {
	var err error
	var action string
	var sessions credentials.Sessions
	var session *credentials.Session
	var role *credentials.Role
	if role, action, err = tui.SelectRolesCredentials(accountAliases); err != nil {
		ExitWithError(12, "failed to cached role credentials", err)
	} else if action != "" {
		return action, role
	}
	if role.Credentials == nil || role.Credentials.IsExpired() {
		if sessions, err = credentials.GetSessions(); err != nil {
			ExitWithError(13, "failed to parse sso sessions", err)
		}
		if session = sessions.FindByName(role.SessionName); session == nil {
			ExitWithError(14, "failed to find sso session "+role.SessionName, err)
		}
		if session.ClientToken == nil || session.ClientToken.IsExpired() {
			if err = tui.ClientLogin(session); err != nil {
				ExitWithError(15, "failed to authorize device login", err)
			}
		}
		if err = session.RefreshRoleCredentials(role); err != nil {
			ExitWithError(16, "failed to get credentials", err)
		}
		if err = role.Credentials.Save(session.Name, role.CacheKey()); err != nil {
			ExitWithError(17, "failed to save credentials", err)
		}
	}
	if err = role.MarkLastUsed(); err != nil {
		ExitWithError(18, "failed to mark last used role", err)
	}
	return "", role
}

func padAccountNumbers(aliases map[string]string) map[string]string {
	result := map[string]string{}
	for account, alias := range aliases {
		result[fmt.Sprintf("%012s", account)] = alias
	}
	return result
}

func setupConfigFile() {
	if homeDir, err := os.UserHomeDir(); err == nil {
		os.MkdirAll(homeDir+"/.aws/knox", os.FileMode(0700))
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetConfigPermissions(os.FileMode(0600))
	viper.AddConfigPath("$HOME/.aws/knox")
	viper.SetDefault("default_connect_uid", uint32(0))
	viper.SetDefault("select_cached_first", false)
	viper.SetDefault("filter_strategy", "fuzzy")
	viper.SetDefault("max_items_to_show", 10)
	viper.SetDefault("account_aliases", map[string]string{})
	viper.SetDefault("instance_col_tags", []string{"Instance Type", "Private IP", "Public IP", "Name"})
	viper.SafeWriteConfig()
	viper.ReadInConfig()
	tui.MaxItemsToShow = viper.GetInt("max_items_to_show")
	tui.FilterStrategy = viper.GetString("filter_strategy")
	selectCachedFirst = viper.GetBool("select_cached_first")
	connectUid = viper.GetUint32("default_connect_uid")
	accountAliases = padAccountNumbers( viper.GetStringMapString("account_aliases") )
	instanceColTags = viper.GetStringSlice("instance_col_tags")

}

func init() {
	RootCmd.Flags().SortFlags = true
	RootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", debug, "Debug mode")

	setupConfigFile()

	if tty, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0); err == nil {
		ansi.Writer = tty
		color.Writer = tty
	}
}
