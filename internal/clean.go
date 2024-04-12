package internal

import (
	"fmt"
	"slices"
	"strings"

	"github.com/null93/aws-knox/sdk/credentials"
	"github.com/spf13/cobra"
)

var (
	cleanAll         = false
	allowedCleanArgs = []string{"creds", "sso"}
)

var cleanCmd = &cobra.Command{
	Use:       "clean [" + strings.Join(allowedCleanArgs, "] [") + "]",
	Short:     "Delete expired role credentials from cache",
	Args:      cobra.RangeArgs(1, 2),
	ValidArgs: allowedCleanArgs,
	Example:   "  knox clean creds\n  knox clean sso -a\n  knox clean creds sso",
	Run: func(cmd *cobra.Command, args []string) {
		if slices.Contains(args, "creds") {
			roles, err := credentials.GetSavedRolesWithCredentials()
			if err != nil {
				ExitWithError(1, "failed to get role credentials", err)
			}
			for _, role := range roles {
				if role.Credentials.IsExpired() || cleanAll {
					err := role.Credentials.DeleteCache(role.CacheKey())
					if err != nil {
						ExitWithError(2, "failed to delete role credentials", err)
					}
				}
			}
			if cleanAll {
				fmt.Println("Successfully deleted all role credentials")
			} else {
				fmt.Println("Successfully deleted expired role credentials")
			}
		}
		if slices.Contains(args, "sso") {
			sessions, err := credentials.GetSessions()
			if err != nil {
				ExitWithError(3, "failed to get configured sessions", err)
			}
			for _, session := range sessions {
				if session.ClientToken != nil && session.ClientToken.IsExpired() || cleanAll {
					err := session.DeleteCache()
					if err != nil {
						ExitWithError(4, "failed to delete client credentials", err)
					}
				}
			}
			if cleanAll {
				fmt.Println("Successfully deleted all client credentials")
			} else {
				fmt.Println("Successfully deleted expired client credentials")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().SortFlags = true
	cleanCmd.Flags().BoolVarP(&cleanAll, "all", "a", cleanAll, "Delete even if not expired credentials")
}
