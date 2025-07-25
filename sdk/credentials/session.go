package credentials

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"gopkg.in/ini.v1"
)

var (
	ErrorRoleCredentialsNil = fmt.Errorf("role credentials are nil")
	ErrRoleNil              = fmt.Errorf("role cannot be nil")
)

type Instances []Instance

type Instance struct {
	Id               string
	InstanceType     string
	PrivateIpAddress string
	PublicIpAddress  string
	Tags             map[string]string
}

func (r *Role) StartSession(instanceId string, defaultUid uint32) (*ssm.StartSessionOutput, error) {
	staticProvider := awscredentials.NewStaticCredentialsProvider(
		r.Credentials.AccessKeyId,
		r.Credentials.SecretAccessKey,
		r.Credentials.SessionToken,
	)
	options := ssm.Options{Region: "us-east-1", Credentials: staticProvider}
	client := ssm.New(options)
	input := ssm.StartSessionInput{
		Target:       &instanceId,
		DocumentName: aws.String("AWS-StartInteractiveCommand"),
		Parameters: map[string][]string{
			"command": {fmt.Sprintf("sudo su - `id -un %d`", defaultUid)},
		},
	}
	return client.StartSession(context.TODO(), &input)
}

func (r *Role) StartCommand(instanceId string, command string) (*ssm.StartSessionOutput, error) {
	staticProvider := awscredentials.NewStaticCredentialsProvider(
		r.Credentials.AccessKeyId,
		r.Credentials.SecretAccessKey,
		r.Credentials.SessionToken,
	)
	options := ssm.Options{Region: "us-east-1", Credentials: staticProvider}
	client := ssm.New(options)
	input := ssm.StartSessionInput{
		Target:       &instanceId,
		DocumentName: aws.String("AWS-StartInteractiveCommand"),
		Parameters: map[string][]string{
			"command": {fmt.Sprintf("sudo -u root bash -c %q", command)},
		},
	}
	return client.StartSession(context.TODO(), &input)
}

func (r *Role) StartPortForward(instanceId string, port, localPort uint16) (*ssm.StartSessionOutput, error) {
	staticProvider := awscredentials.NewStaticCredentialsProvider(
		r.Credentials.AccessKeyId,
		r.Credentials.SecretAccessKey,
		r.Credentials.SessionToken,
	)
	options := ssm.Options{Region: "us-east-1", Credentials: staticProvider}
	client := ssm.New(options)
	input := ssm.StartSessionInput{
		Target:       &instanceId,
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters: map[string][]string{
			"portNumber":      []string{fmt.Sprintf("%d", port)},
			"localPortNumber": []string{fmt.Sprintf("%d", localPort)},
		},
	}
	return client.StartSession(context.TODO(), &input)
}

type Roles []Role

type Role struct {
	Name        string           `json:"name"`
	AccountId   string           `json:"accountId"`
	Region      string           `json:"region"`
	SessionName string           `json:"sessionName"`
	Credentials *RoleCredentials `json:"-"`
}

func (r *Role) CacheKey() string {
	return r.Region + "_" + r.AccountId + "_" + r.Name
}

func (r Roles) FindByName(name string) *Role {
	for _, role := range r {
		if role.Name == name {
			return &role
		}
	}
	return nil
}

func (r *Role) GetManagedInstances(region string) (Instances, error) {
	instances := Instances{}
	if r.Credentials == nil {
		return instances, ErrorRoleCredentialsNil
	}
	staticProvider := awscredentials.NewStaticCredentialsProvider(
		r.Credentials.AccessKeyId,
		r.Credentials.SecretAccessKey,
		r.Credentials.SessionToken,
	)
	options := ec2.Options{Region: region, Credentials: staticProvider}
	client := ec2.New(options)
	params := ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	}
	paginator := ec2.NewDescribeInstancesPaginator(client, &params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return instances, err
		}
		for _, info := range page.Reservations {
			for _, instance := range info.Instances {
				privateId := "-"
				publicId := "-"
				if instance.PrivateIpAddress != nil {
					privateId = aws.ToString(instance.PrivateIpAddress)
				}
				if instance.PublicIpAddress != nil {
					publicId = aws.ToString(instance.PublicIpAddress)
				}
				tags := map[string]string{}
				for _, tag := range instance.Tags {
					key := aws.ToString(tag.Key)
					value := aws.ToString(tag.Value)
					tags[key] = value
				}
				instance := Instance{
					Id:               aws.ToString(instance.InstanceId),
					InstanceType:     string(instance.InstanceType),
					PrivateIpAddress: privateId,
					PublicIpAddress:  publicId,
					Tags:             tags,
				}
				instances = append(instances, instance)
			}
		}
	}
	return instances, nil
}

type Accounts []Account

type Account struct {
	Id    string
	Email string
	Name  string
}

func (a Accounts) FindById(id string) *Account {
	for _, account := range a {
		if account.Id == id {
			return &account
		}
	}
	return nil
}

type Session struct {
	Name              string
	Region            string
	StartUrl          string
	Scopes            []string
	ClientCredentials *ClientCredentials
	ClientToken       *ClientToken
}

type Sessions []Session

func (s Sessions) FindByName(name string) *Session {
	for _, session := range s {
		if session.Name == name {
			return &session
		}
	}
	return nil
}

func GetSessions() (Sessions, error) {
	sessions := Sessions{}
	homePath, err := os.UserHomeDir()
	if err != nil {
		return sessions, err
	}
	awsConfigPath := filepath.Join(homePath, ".aws", "config")
	config, err := ini.Load(awsConfigPath)
	if err != nil {
		return sessions, err
	}
	for _, section := range config.Sections() {
		name := section.Name()
		if strings.HasPrefix(name, "sso-session ") {
			session := Session{
				Name:     strings.TrimPrefix(name, "sso-session "),
				Region:   section.Key("sso_region").String(),
				StartUrl: section.Key("sso_start_url").String(),
				Scopes:   strings.Split(section.Key("sso_registration_scopes").String(), ","),
			}
			cachedCredentials, err := session.findClientCredentials()
			if err != nil {
				return sessions, err
			}
			cachedToken, err := session.findClientToken()
			if err != nil {
				return sessions, err
			}
			session.ClientCredentials = cachedCredentials
			session.ClientToken = cachedToken
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func fileSafeKey(key string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(key))
	return hex.EncodeToString(sha1.Sum(nil))
}

func (s *Session) clientCredentialsCacheKey() string {
	format := `{"region": "%s", "scopes": [%s], "session_name": "%s", "startUrl": "%s", "tool": "botocore"}`
	serializedScopes := `"` + strings.Join(s.Scopes, `", "`) + `"`
	key := fmt.Sprintf(format, s.Region, serializedScopes, s.Name, s.StartUrl)
	return fileSafeKey(key)
}

func (s *Session) clientTokenCacheKey() string {
	return fileSafeKey(s.Name)
}

func (s *Session) Save() error {
	if s.ClientCredentials != nil {
		key := s.clientCredentialsCacheKey()
		if err := s.ClientCredentials.save(key); err != nil {
			return err
		}
	}
	if s.ClientToken != nil {
		key := s.clientTokenCacheKey()
		if err := s.ClientToken.save(key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) DeleteCache() error {
	if s.ClientCredentials != nil {
		key := s.clientCredentialsCacheKey()
		if err := s.ClientCredentials.delete(key); err != nil {
			return err
		}
	}
	if s.ClientToken != nil {
		key := s.clientTokenCacheKey()
		if err := s.ClientToken.delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) RegisterClient() error {
	options := ssooidc.Options{Region: s.Region}
	client := ssooidc.New(options)
	register, err := client.RegisterClient(context.TODO(), &ssooidc.RegisterClientInput{
		ClientName: aws.String("knox-client-" + s.Name),
		ClientType: aws.String("public"),
		Scopes:     s.Scopes,
	})
	if err != nil {
		return err
	}
	s.ClientCredentials = &ClientCredentials{
		ClientId:     *register.ClientId,
		ClientSecret: *register.ClientSecret,
		ExpiresAt:    time.Unix(register.ClientSecretExpiresAt, 0).UTC(),
		Scopes:       s.Scopes,
	}
	return nil
}

func (s *Session) StartDeviceAuthorization() (string, string, string, string, error) {
	options := ssooidc.Options{Region: s.Region}
	client := ssooidc.New(options)
	deviceAuth, err := client.StartDeviceAuthorization(context.TODO(), &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     &s.ClientCredentials.ClientId,
		ClientSecret: &s.ClientCredentials.ClientSecret,
		StartUrl:     &s.StartUrl,
	})
	if err != nil {
		return "", "", "", "", err
	}
	userCode := aws.ToString(deviceAuth.UserCode)
	deviceCode := aws.ToString(deviceAuth.DeviceCode)
	url := aws.ToString(deviceAuth.VerificationUri)
	urlFull := aws.ToString(deviceAuth.VerificationUriComplete)
	return userCode, deviceCode, url, urlFull, nil
}

func (s *Session) WaitForToken(deviceCode string) error {
	options := ssooidc.Options{Region: s.Region}
	client := ssooidc.New(options)
	token := &ssooidc.CreateTokenOutput{}
	var err error
	for {
		token, err = client.CreateToken(context.TODO(), &ssooidc.CreateTokenInput{
			ClientId:     aws.String(s.ClientCredentials.ClientId),
			ClientSecret: aws.String(s.ClientCredentials.ClientSecret),
			DeviceCode:   aws.String(deviceCode),
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		})
		if err != nil {
			if strings.Contains(err.Error(), "AuthorizationPendingException") {
				time.Sleep(1 * time.Second)
				continue
			}
			return err
		}
		break
	}
	s.ClientToken = &ClientToken{
		AccessToken:           aws.ToString(token.AccessToken),
		ClientId:              s.ClientCredentials.ClientId,
		ClientSecret:          s.ClientCredentials.ClientSecret,
		ExpiresAt:             time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).UTC(),
		RefreshToken:          aws.ToString(token.RefreshToken),
		Region:                s.Region,
		RegistrationExpiresAt: s.ClientCredentials.ExpiresAt,
		StartUrl:              s.StartUrl,
	}
	return nil
}

func (s *Session) RefreshToken() error {
	options := ssooidc.Options{Region: s.Region}
	client := ssooidc.New(options)
	token, err := client.CreateToken(context.TODO(), &ssooidc.CreateTokenInput{
		ClientId:     aws.String(s.ClientCredentials.ClientId),
		ClientSecret: aws.String(s.ClientCredentials.ClientSecret),
		RefreshToken: aws.String(s.ClientToken.RefreshToken),
		GrantType:    aws.String("refresh_token"),
	})
	if err != nil {
		return err
	}
	s.ClientToken = &ClientToken{
		AccessToken:           aws.ToString(token.AccessToken),
		ClientId:              s.ClientCredentials.ClientId,
		ClientSecret:          s.ClientCredentials.ClientSecret,
		ExpiresAt:             time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).UTC(),
		RefreshToken:          aws.ToString(token.RefreshToken),
		Region:                s.Region,
		RegistrationExpiresAt: s.ClientCredentials.ExpiresAt,
		StartUrl:              s.StartUrl,
	}
	return nil
}

func (s *Session) GetAccountsStream(onAccounts func([]Account)) error {
	options := sso.Options{Region: s.Region}
	client := sso.New(options)
	params := sso.ListAccountsInput{AccessToken: &s.ClientToken.AccessToken}
	paginator := sso.NewListAccountsPaginator(client, &params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		accounts := []Account{}
		for _, details := range page.AccountList {
			account := Account{
				Id:    aws.ToString(details.AccountId),
				Email: aws.ToString(details.EmailAddress),
				Name:  aws.ToString(details.AccountName),
			}
			accounts = append(accounts, account)
		}
		if len(accounts) > 0 && onAccounts != nil {
			onAccounts(accounts)
		}
	}
	return nil
}

func (s *Session) GetRoles(accountId string) (Roles, error) {
	roles := Roles{}
	options := sso.Options{Region: s.Region}
	client := sso.New(options)
	params := sso.ListAccountRolesInput{AccessToken: &s.ClientToken.AccessToken, AccountId: &accountId}
	paginator := sso.NewListAccountRolesPaginator(client, &params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return roles, err
		}
		for _, details := range page.RoleList {
			roleName := aws.ToString(details.RoleName)
			role := Role{
				Name:        roleName,
				AccountId:   accountId,
				Region:      s.Region,
				SessionName: s.Name,
			}
			creds, err := findRoleCredentials(role)
			if err != nil {
				return roles, err
			}
			role.Credentials = creds
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (s *Session) RefreshRoleCredentials(role *Role) error {
	if role == nil {
		return ErrRoleNil
	}
	options := sso.Options{Region: s.Region}
	client := sso.New(options)
	params := sso.GetRoleCredentialsInput{
		AccessToken: &s.ClientToken.AccessToken,
		AccountId:   &role.AccountId,
		RoleName:    &role.Name,
	}
	resp, err := client.GetRoleCredentials(context.TODO(), &params)
	if err != nil {
		return err
	}
	role.Credentials = &RoleCredentials{
		Version:         1,
		AccessKeyId:     aws.ToString(resp.RoleCredentials.AccessKeyId),
		SecretAccessKey: aws.ToString(resp.RoleCredentials.SecretAccessKey),
		SessionToken:    aws.ToString(resp.RoleCredentials.SessionToken),
		Expiration:      time.Unix(resp.RoleCredentials.Expiration/1000, 0),
	}
	return nil
}
