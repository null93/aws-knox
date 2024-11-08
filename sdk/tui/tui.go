package tui

import (
	"fmt"
	"strings"
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
	attemptReauth := false
	if session.ClientCredentials != nil && !session.ClientCredentials.IsExpired() {
		// Try to refresh. If it works, then it works. If it fails, then we need to re-auth.
		err := session.RefreshToken()
		if err != nil {
			attemptReauth = true
		}
	}
	if session.ClientCredentials == nil || session.ClientCredentials.IsExpired() || attemptReauth {
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
	p.WithHeaders("SSO Session", "Region", "SSO Start URL", "Expires In")
	p.AddAction(keys.Tab, "tab", "view cached")
	for _, session := range sessions {
		expires := "-"
		if session.ClientToken != nil && !session.ClientToken.IsExpired() {
			expires = fmt.Sprintf("%.f mins", session.ClientToken.ExpiresAt.Sub(now).Minutes())
		}
		p.AddOption(session.Name, session.Name, session.Region, session.StartUrl, expires)
	}
	selection, firedKeyCode := p.Pick("")
	if firedKeyCode != nil && *firedKeyCode == keys.Tab {
		return "", "toggle-view", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedSession
	}
	return selection.Value.(string), "", nil
}

func SelectAccount(session *credentials.Session, accountAliases map[string]string) (string, string, error) {
	accountIds, err := session.GetAccounts()
	if err != nil {
		return "", "", err
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Accounts Found")
	p.WithTitle("Pick Account")
	p.WithHeaders("Account ID", "Alias/Name", "Email")
	p.AddAction(keys.Esc, "esc", "go back")
	for _, account := range accountIds {
		name := account.Name
		if val, ok := accountAliases[account.Id]; ok {
			if strings.TrimSpace(val) != "" {
				name = val
			}
		}
		p.AddOption(account.Id, account.Id, name, account.Email)
	}
	selection, firedKeyCode := p.Pick("")
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
	selection, firedKeyCode := p.Pick("")
	if firedKeyCode != nil && *firedKeyCode == keys.Esc {
		return "", "back", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedRole
	}
	return selection.Value.(string), "", nil
}

func cutOff(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

func cutOffInMiddle(s string, n int) string {
	l := len(s)
	if l > n {
		return s[:n/2] + "…" + s[l-(n/2):]
	}
	return s
}

func SelectInstance(role *credentials.Role, region, initialFilter string, instanceColTags []string) (string, string, error) {
	instances, err := role.GetManagedInstances(region)
	if err != nil {
		return "", "", err
	}
	cols := []string{"Instance ID"}
	for _, tag := range instanceColTags {
		cols = append(cols, tag)
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Instances Found")
	p.WithTitle(fmt.Sprintf("Pick EC2 Instance (%s)", region))
	p.WithHeaders(cols...)
	p.AddAction(keys.Esc, "esc", "go back")
	p.AddAction(keys.F1, "f1", "pick region")
	for _, instance := range instances {
		values := []string{instance.Id}
		for _, tag := range instanceColTags {
			value := "-"
			switch tag {
			case "Instance Type":
				value = instance.InstanceType
			case "Private IP":
				value = instance.PrivateIpAddress
			case "Public IP":
				value = instance.PublicIpAddress
			default:
				value = instance.Tags[tag]
			}
			value = cutOffInMiddle(value, 36)
			values = append(values, value)
		}
		p.AddOption(instance.Id, values...)
	}
	selection, firedKeyCode := p.Pick(initialFilter)
	if firedKeyCode != nil && *firedKeyCode == keys.Esc {
		return "", "back", nil
	}
	if firedKeyCode != nil && *firedKeyCode == keys.F1 {
		return "", "pick-region", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedInstance
	}
	return selection.Value.(string), "", nil
}

func SelectRegion(initialFilter string) (string, string, error) {
	regions := [][]string{
		{"us-east-1", "US East (N. Virginia)"},
		{"us-east-2", "US East (Ohio)"},
		{"us-west-1", "US West (N. California)"},
		{"us-west-2", "US West (Oregon)"},
		{"af-south-1", "Africa (Cape Town)"},
		{"ap-east-1", "Asia Pacific (Hong Kong)"},
		{"ap-south-1", "Asia Pacific (Mumbai)"},
		{"ap-southeast-1", "Asia Pacific (Singapore)"},
		{"ap-southeast-2", "Asia Pacific (Sydney)"},
		{"ap-northeast-1", "Asia Pacific (Tokyo)"},
		{"ap-northeast-2", "Asia Pacific (Seoul)"},
		{"ca-central-1", "Canada (Central)"},
		{"eu-central-1", "EU (Frankfurt)"},
		{"eu-west-1", "EU (Ireland)"},
		{"eu-west-2", "EU (London)"},
		{"eu-south-1", "EU (Milan)"},
		{"eu-west-3", "EU (Paris)"},
		{"eu-north-1", "EU (Stockholm)"},
		{"me-south-1", "Middle East (Bahrain)"},
		{"sa-east-1", "South America (Sao Paulo)"},
		{"ap-northeast-3", "Asia Pacific (Osaka-Local)"},
		{"us-gov-west-1", "AWS GovCloud (US-West)"},
		{"cn-northwest-1", "China (Ningxia)"},
		{"cn-north-1", "China (Beijing)"},
		{"us-gov-east-1", "AWS GovCloud (US-East)"},
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Regions Found")
	p.WithTitle("Pick Region")
	p.WithHeaders("Region", "Name")
	p.AddAction(keys.Esc, "esc", "go back")
	for _, regionArray := range regions {
		p.AddOption(regionArray[0], regionArray[0], regionArray[1])
	}
	selection, firedKeyCode := p.Pick("")
	if firedKeyCode != nil && *firedKeyCode == keys.Esc {
		return "", "back", nil
	}
	if selection == nil {
		return "", "", ErrNotPickedInstance
	}
	return selection.Value.(string), "", nil
}

func SelectRolesCredentials(accountAliases map[string]string) (*credentials.Role, string, error) {
	now := time.Now()
	roles, err := credentials.GetSavedRolesWithCredentials()
	if err != nil {
		return nil, "", err
	}
	p := picker.NewPicker()
	p.WithMaxHeight(MaxItemsToShow)
	p.WithEmptyMessage("No Role Credentials Found")
	p.WithTitle("Pick Role Credentials")
	p.WithHeaders("SSO Session", "Region", "Account ID", "Alias", "Role Name", "Expires In")
	p.AddAction(keys.Tab, "tab", "pick session")
	p.AddAction(keys.Delete, "del", "delete")
	for _, role := range roles {
		expires := "-"
		if role.Credentials != nil && !role.Credentials.IsExpired() {
			expires = fmt.Sprintf("%.f mins", role.Credentials.Expiration.Sub(now).Minutes())
		}
		alias := "-"
		if val, ok := accountAliases[role.AccountId]; ok {
			if strings.TrimSpace(val) != "" {
				alias = val
			}
		}
		p.AddOption(role, role.SessionName, role.Region, role.AccountId, alias, role.Name, expires)
	}
	selection, firedKeyCode := p.Pick("")
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
