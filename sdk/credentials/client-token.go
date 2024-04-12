package credentials

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	ClientTokenCachePath = ".aws/sso/cache"
)

type ClientToken struct {
	AccessToken           string    `json:"accessToken"`
	ClientId              string    `json:"clientId"`
	ClientSecret          string    `json:"clientSecret"`
	ExpiresAt             time.Time `json:"expiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	Region                string    `json:"region"`
	RegistrationExpiresAt time.Time `json:"registrationExpiresAt"`
	StartUrl              string    `json:"startUrl"`
}

func (t *ClientToken) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}

func (s *Session) findClientToken() (*ClientToken, error) {
	cacheKey := s.clientTokenCacheKey()
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(homedir, ClientTokenCachePath, cacheKey+".json")
	if _, err := os.Stat(cachePath); err == nil {
		contents, err := ioutil.ReadFile(cachePath)
		if err != nil {
			return nil, err
		}
		token := &ClientToken{}
		if err := json.Unmarshal(contents, token); err != nil {
			return nil, err
		}
		if !token.IsExpired() {
			return token, nil
		}
		return token, nil
	}
	return nil, nil
}

func (t *ClientToken) save(key string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(homedir, ClientTokenCachePath), 0700); err != nil {
		return err
	}
	cachePath := filepath.Join(homedir, ClientTokenCachePath, key+".json")
	contents, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cachePath, contents, 0600)
}

func (t *ClientToken) delete(key string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cachePath := filepath.Join(homedir, ClientTokenCachePath, key+".json")
	return os.Remove(cachePath)
}
