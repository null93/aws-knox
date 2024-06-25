package internal

import (
	"fmt"
	"os"
	"syscall"

	"github.com/null93/aws-knox/pkg/ansi"
	"github.com/null93/aws-knox/pkg/color"
	"github.com/null93/aws-knox/sdk/credentials"
	. "github.com/null93/aws-knox/sdk/style"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	Version = "0.0.0"
	Debug   = false
)

var RootCmd = &cobra.Command{
	Use:     "knox",
	Version: Version,
	Short:   "knox helps you manage AWS credentials that are created via AWS SSO",
}

func ExitWithError(code int, message string, err error) {
	fmt.Printf("Error: %s\n", message)
	if err != nil && Debug {
		fmt.Printf("Debug: %s\n", err.Error())
	}
	os.Exit(code)
}

func ClientLogin(session *credentials.Session) error {
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
	err = session.Save()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	RootCmd.Flags().SortFlags = true
	RootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	RootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", Debug, "Debug mode")

	if tty, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0); err == nil {
		ansi.Writer = tty
		color.Writer = tty
	}
}
