package main

import (
	"fmt"
	"time"

	"github.com/null93/aws-knox/sdk/credentials"
)

func main() {
	sessions, _ := credentials.GetSessions()
	for _, session := range sessions {
		fmt.Println("Session Name:      ", session.Name)
		fmt.Println("Region:            ", session.Region)
		fmt.Println("Start URL:         ", session.StartUrl)
		fmt.Println("Scopes:            ", session.Scopes)
		fmt.Println("Client Credentials:", session.ClientCredentials != nil)
		fmt.Println("Client Token:      ", session.ClientToken != nil)
		fmt.Println()
	}
	session := sessions.FindByName("staging-sso")
	fmt.Println(session.Name)
	fmt.Println(session.ClientToken.ExpiresAt)
	fmt.Println(session.ClientToken.IsExpired())
	fmt.Println(time.Now().UTC())
}
