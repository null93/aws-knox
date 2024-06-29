package internal

import (
	"fmt"
	"os"
	"syscall"

	"github.com/null93/aws-knox/pkg/ansi"
	"github.com/null93/aws-knox/pkg/color"
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
