package tui

import (
	"fmt"
	"time"

	"atomicgo.dev/keyboard/keys"
	"github.com/null93/aws-knox/pkg/ansi"
	"github.com/null93/aws-knox/pkg/color"
	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/null93/aws-knox/sdk/picker"
	. "github.com/null93/aws-knox/sdk/style"
	"github.com/pkg/browser"
)

var (
	MaxItemsToShow              int = 10
	ErrNotPickedSession             = fmt.Errorf("no sso session picked")
	ErrNotPickedAccount             = fmt.Errorf("no account picked")
	ErrNotPickedRole                = fmt.Errorf("no role picked")
	ErrNotPickedInstance            = fmt.Errorf("no instance picked")
	ErrNotPickedRoleCredentials     = fmt.Errorf("no role credentials picked")
)

func ClientLogin(session *credentials.Session) error {
	if session.ClientCredentials != nil && !session.ClientCredentials.IsExpired() {
		session.RefreshToken()
	}
	if session.ClientCredentials == nil || session.ClientCredentials.IsExpired() {
		if err := session.RegisterClient(); err != nil {
			return err
		}
		userCode, deviceCode, url, urlFull, err := session.StartDeviceAuthorization()
		if err != nil {
			return err
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
			return err
		}
		err = session.WaitForToken(deviceCode)
		ansi.MoveCursorUp(6)
		ansi.ClearDown()
		if err != nil {
			return err
		}
	}
	if err := session.Save(); err != nil {
		return err
	}
	return nil
}

func SelectSession(sessions credentials.Sessions) (string, string, error) {
	now := time.Now()
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No SSO Sessions Found")
	p.WithTitle("Pick SSO Session")
	p.WithHeaders("SSO Session", "Region", "SSO Start URL", "Refreshable", "Expires In")
	p.AddAction(keys.Tab, "tab", "view cached")
	for _, session := range sessions {
		expires := "-"
		refreshable := "-"
		if session.ClientToken != nil && !session.ClientToken.IsExpired() {
			expires = fmt.Sprintf("%.f mins", session.ClientToken.ExpiresAt.Sub(now).Minutes())
		}
		if session.ClientCredentials != nil && !session.ClientCredentials.IsExpired() {
			hours := session.ClientCredentials.ExpiresAt.Sub(now).Hours()
			if hours < 1 {
				refreshable = fmt.Sprintf("%.f mins", hours * 60)
			} else if hours < 24 {
				refreshable = fmt.Sprintf("%.f hours", hours)
			} else {
				refreshable = fmt.Sprintf("%.f days", hours / 24)
			}
		}
		p.AddOption(session.Name, session.Name, session.Region, session.StartUrl, refreshable, expires)
	}
	selection, firedKeyCode := p.Pick()
	if firedKeyCode != nil && *firedKeyCode == keys.Tab {
		return "", "toggle-view", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedSession
	}
	return selection.Value.(string), "", nil
}

func SelectAccount(session *credentials.Session) (string, string, error) {
	accountIds, err := session.GetAccounts()
	if err != nil {
		return "", "", err
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Accounts Found")
	p.WithTitle("Pick Account")
	p.WithHeaders("Account ID", "Name", "Email")
	p.AddAction(keys.Esc, "esc", "go back")
	for _, account := range accountIds {
		p.AddOption(account.Id, account.Id, account.Name, account.Email)
	}
	selection, firedKeyCode := p.Pick()
	if firedKeyCode != nil && *firedKeyCode == keys.Esc {
		return "", "back", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedAccount
	}
	return selection.Value.(string), "", nil
}

func SelectRole(roles credentials.Roles) (string, string, error) {
	now := time.Now()
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Roles Found")
	p.WithTitle("Pick Role")
	p.WithHeaders("Role Name", "Expires In")
	p.AddAction(keys.Esc, "esc", "go back")
	for _, role := range roles {
		expires := "-"
		if role.Credentials != nil && !role.Credentials.IsExpired() {
			expires = fmt.Sprintf("%.f mins", role.Credentials.Expiration.Sub(now).Minutes())
		}
		p.AddOption(role.Name, role.Name, expires)
	}
	selection, firedKeyCode := p.Pick()
	if firedKeyCode != nil && *firedKeyCode == keys.Esc {
		return "", "back", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedRole
	}
	return selection.Value.(string), "", nil
}

func SelectInstance(role *credentials.Role) (string, string, error) {
	instances, err := role.GetManagedInstances()
	if err != nil {
		return "", "", err
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Instances Found")
	p.WithTitle("Pick EC2 Instance")
	p.WithHeaders("Instance ID", "Instance Type", "Private IP", "Public IP", "Name")
	p.AddAction(keys.Esc, "esc", "go back")
	for _, instance := range instances {
		p.AddOption(instance.Id, instance.Id, instance.InstanceType, instance.PrivateIpAddress, instance.PublicIpAddress, instance.Name)
	}
	selection, firedKeyCode := p.Pick()
	if firedKeyCode != nil && *firedKeyCode == keys.Esc {
		return "", "back", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedInstance
	}
	return selection.Value.(string), "", nil
}

func SelectRolesCredentials() (*credentials.Role, string, error) {
	now := time.Now()
	roles, err := credentials.GetSavedRolesWithCredentials()
	if err != nil {
		return nil, "", err
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Role Credentials Found")
	p.WithTitle("Pick Role Credentials")
	p.WithHeaders("SSO Session", "Region", "Account ID", "Role Name", "Expires In")
	p.AddAction(keys.Tab, "tab", "pick session")
	p.AddAction(keys.Delete, "del", "delete")
	for _, role := range roles {
		expires := "-"
		if role.Credentials != nil && !role.Credentials.IsExpired() {
			expires = fmt.Sprintf("%.f mins", role.Credentials.Expiration.Sub(now).Minutes())
		}
		p.AddOption(role, role.SessionName, role.Region, role.AccountId, role.Name, expires)
	}
	selection, firedKeyCode := p.Pick()
	if firedKeyCode != nil && *firedKeyCode == keys.Tab {
		return nil, "toggle-view", nil
	}
	if firedKeyCode != nil && *firedKeyCode == keys.Delete {
		if selection != nil {
			selected := selection.Value.(credentials.Role)
			return &selected, "delete", nil
		}
		return nil, "delete", nil
	}
	if selection == nil {
		return nil, "", ErrNotPickedRoleCredentials
	}
	selected := selection.Value.(credentials.Role)
	return &selected, "", nil
}
