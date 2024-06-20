package internal

import (
	"fmt"
	"time"

	"github.com/null93/aws-knox/pkg/ansi"
	"github.com/null93/aws-knox/pkg/color"
	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/picker"
	. "github.com/null93/aws-knox/sdk/style"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	sessionName string
	accountId   string
	roleName    string
)

var selectCredentialsCmd = &cobra.Command{
	Use:   "select",
	Short: "Select specific AWS role credentials starting from SSO",
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
		if sessionName == "" {
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
			sessionName = selection.Value.(string)
		}
		session := sessions.FindByName(sessionName)
		if session == nil {
			ExitWithError(4, "session with passed name not found", err)
		}
		if session.ClientToken == nil || session.ClientToken.IsExpired() {
			if err := session.RegisterClient(); err != nil {
				ExitWithError(5, "failed to register client", err)
			}
			userCode, deviceCode, url, urlFull, err := session.StartDeviceAuthorization()
			if err != nil {
				ExitWithError(6, "failed to start device authorization", err)
			}
			yellow := color.ToForeground(YellowColor).Decorator()
			gray := color.ToForeground(LightGrayColor).Decorator()
			title := TitleStyle.Decorator()
			DefaultStyle.Printfln("")
			DefaultStyle.Printfln("%s %s", title("SSO Session:      "), gray(session.Name))
			DefaultStyle.Printfln("%s %s", title("SSO Start URL:    "), gray(session.StartUrl))
			DefaultStyle.Printfln("%s %s", title("Authorization URL:"), gray(url))
			DefaultStyle.Printfln("%s %s", title("Device Code:      "), yellow(userCode))
			DefaultStyle.Printfln("")
			DefaultStyle.Printf("Waiting for authorization to complete...")
			err = browser.OpenURL(urlFull)
			if err != nil {
				ansi.MoveCursorUp(6)
				ansi.ClearDown()
				ExitWithError(7, "failed to open url in browser", err)
			}
			err = session.WaitForToken(deviceCode)
			ansi.MoveCursorUp(6)
			ansi.ClearDown()
			if err != nil {
				ExitWithError(8, "failed to wait for token", err)
			}
			err = session.Save()
			if err != nil {
				ExitWithError(9, "failed to save session", err)
			}
		}
		if accountId == "" {
			accountIds, err := session.GetAccounts()
			if err != nil {
				ExitWithError(10, "failed to get account ids", err)
			}
			if len(accountIds) == 0 {
				ExitWithError(11, "no accounts found", err)
			}
			p := picker.NewPicker()
			p.WithMaxHeight(5)
			p.WithEmptyMessage("No Accounts Found")
			p.WithTitle("Pick Account")
			p.WithHeaders("Account ID", "Name", "Email")
			for _, account := range accountIds {
				p.AddOption(account.Id, account.Id, account.Name, account.Email)
			}
			selection := p.Pick()
			if selection == nil {
				ExitWithError(12, "failed to pick an account id", err)
			}
			accountId = selection.Value.(string)
		}
		roles, err := session.GetRoles(accountId)
		if roleName == "" {
			if err != nil {
				ExitWithError(13, "failed to get roles", err)
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
				ExitWithError(14, "failed to pick a role name", err)
			}
			roleName = selection.Value.(string)
		}
		role := roles.FindByName(roleName)
		if role == nil {
			ExitWithError(15, "role with passed name not found", err)
		}
		if role.Credentials == nil || role.Credentials.IsExpired() {
			err := session.RefreshRoleCredentials(role)
			if err != nil {
				ExitWithError(16, "failed to get credentials", err)
			}
			err = role.Credentials.Save(session.Name, role.CacheKey())
			if err != nil {
				ExitWithError(17, "failed to save credentials", err)
			}
		}
		if err := role.MarkLastUsed(); err != nil {
			ExitWithError(18, "failed to mark last used role", err)
		}
		json, err := role.Credentials.ToJSON()
		if err != nil {
			ExitWithError(19, "failed to convert credentials to json", err)
		}
		fmt.Println(json)
	},
}

func init() {
	RootCmd.AddCommand(selectCredentialsCmd)
	selectCredentialsCmd.Flags().SortFlags = true
	selectCredentialsCmd.Flags().StringVarP(&sessionName, "sso-session", "s", sessionName, "SSO session name")
	selectCredentialsCmd.Flags().StringVarP(&accountId, "account-id", "a", accountId, "AWS account ID")
	selectCredentialsCmd.Flags().StringVarP(&roleName, "role-name", "r", roleName, "AWS role name")
}
