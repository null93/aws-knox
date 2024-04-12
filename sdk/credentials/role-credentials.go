package credentials

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RoleCredentials struct {
	Version         uint      `json:"Version"`
	AccessKeyId     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      time.Time `json:"Expiration"`
}

const (
	KnoxPath                 = ".aws/knox"
	RoleCredentialsCachePath = ".aws/knox/cache"
)

func (r *RoleCredentials) ToJSON() (string, error) {
	contents, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func findRoleCredentials(r Role) (*RoleCredentials, error) {
	cacheKey := r.CacheKey()
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(homedir, RoleCredentialsCachePath, cacheKey+".json")
	if _, err := os.Stat(cachePath); err == nil {
		contents, err := ioutil.ReadFile(cachePath)
		if err != nil {
			return nil, err
		}
		creds := &RoleCredentials{}
		if err := json.Unmarshal(contents, creds); err != nil {
			return nil, err
		}
		return creds, nil
	}
	return nil, nil
}

func (r *RoleCredentials) IsExpired() bool {
	return r.Expiration.Before(time.Now())
}

func (r *RoleCredentials) Save(key string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(homedir, RoleCredentialsCachePath), 0700); err != nil {
		return err
	}
	cachePath := filepath.Join(homedir, RoleCredentialsCachePath, key+".json")
	contents, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cachePath, contents, 0600)
}

func (r *RoleCredentials) DeleteCache(key string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cachePath := filepath.Join(homedir, RoleCredentialsCachePath, key+".json")
	return os.Remove(cachePath)
}

func (r *Role) MarkLastUsed() error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(homedir, KnoxPath), 0700); err != nil {
		return err
	}
	lastUsedPath := filepath.Join(homedir, KnoxPath, "last-used")
	return ioutil.WriteFile(lastUsedPath, []byte(r.CacheKey()), 0600)
}

func GetLastUsedRole() (Role, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return Role{}, err
	}
	lastUsedPath := filepath.Join(homedir, KnoxPath, "last-used")
	contents, err := ioutil.ReadFile(lastUsedPath)
	if err != nil {
		return Role{}, err
	}
	parts := strings.Split(string(contents), "_")
	if len(parts) != 3 {
		return Role{}, fmt.Errorf("invalid last used role")
	}
	region := parts[0]
	accountId := parts[1]
	roleName := parts[2]
	role := Role{
		Region:    region,
		AccountId: accountId,
		Name:      roleName,
	}
	creds, err := findRoleCredentials(role)
	if err != nil {
		return Role{}, err
	}
	role.Credentials = creds
	return role, nil
}

func GetSavedRolesWithCredentials() (Roles, error) {
	roles := Roles{}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return roles, err
	}
	cacheDir := filepath.Join(homedir, RoleCredentialsCachePath)
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return roles, err
	}
	for _, file := range files {
		filename := file.Name()
		if !file.IsDir() && filepath.Ext(filename) == ".json" {
			contents, err := os.ReadFile(filepath.Join(cacheDir, filename))
			parts := strings.Split(filename, "_")
			if len(parts) != 3 {
				continue
			}
			region := parts[0]
			accountId := parts[1]
			roleName := strings.TrimSuffix(parts[2], ".json")
			if err != nil {
				return nil, err
			}
			cred := RoleCredentials{}
			if err := json.Unmarshal(contents, &cred); err != nil {
				return nil, err
			}
			role := Role{
				Region:      region,
				AccountId:   accountId,
				Name:        roleName,
				Credentials: &cred,
			}
			roles = append(roles, role)
		}
	}
	return roles, nil
}
