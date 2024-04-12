package credentials

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	ClientCredentialsCachePath = ".aws/sso/cache"
)

type ClientCredentials struct {
	ClientId     string    `json:"clientId"`
	ClientSecret string    `json:"clientSecret"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Scopes       []string  `json:"scopes"`
}

func (c *ClientCredentials) IsExpired() bool {
	return c.ExpiresAt.Before(time.Now())
}

func (s *Session) findClientCredentials() (*ClientCredentials, error) {
	key := s.clientCredentialsCacheKey()
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(homedir, ClientCredentialsCachePath, key+".json")
	if _, err := os.Stat(cachePath); err == nil {
		contents, err := ioutil.ReadFile(cachePath)
		if err != nil {
			return nil, err
		}
		credentials := &ClientCredentials{}
		if err := json.Unmarshal(contents, credentials); err != nil {
			return nil, err
		}
		if !credentials.IsExpired() {
			return credentials, nil
		}
	}
	return nil, nil
}

func (c *ClientCredentials) save(key string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(homedir, ClientCredentialsCachePath), 0700); err != nil {
		return err
	}
	cachePath := filepath.Join(homedir, ClientCredentialsCachePath, key+".json")
	contents, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cachePath, contents, 0600)
}

func (c *ClientCredentials) delete(key string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cachePath := filepath.Join(homedir, ClientCredentialsCachePath, key+".json")
	return os.Remove(cachePath)
}
