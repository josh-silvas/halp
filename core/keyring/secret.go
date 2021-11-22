package keyring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/josh5276/keyring"
)

// Credential type is used as the type set/retrieved when
// interacting with the gokeys package
type Credential struct {
	Username string
	Password string
	Expire   int64
}

// getCredential function will take a username and service type to get a user credential. The
// expire timer is required here to determine if the password returned from the keyring service
// is actually valid.
func (s *Settings) getCredential(user string, service Service) (Credential, error) {
	cr := Credential{Username: user}
	if _, ok := s.Key[service]; !ok {
		return cr, fmt.Errorf("%s does not exist in the keyring store", service.Name)
	}
	key, err := s.Key[service].Get(svcUser(user, service.Name))
	if err != nil {
		return cr, fmt.Errorf("service.Get:%s", err)
	}
	parsed, err := parseCredential(key)
	if err != nil {
		return cr, fmt.Errorf("parseCredential:%s", err)
	}
	if !parsed.isExpired() {
		parsed.Username = user
		return parsed, nil
	}
	return parsed, errors.New("password is expired")
}

// setCredential function is a small wrapper to the keyring Set function, but with the
// formatting of the password to match this packages expire types.
func (s *Settings) setCredential(user, key string, service Service, expire int64) (Credential, error) {
	cred := Credential{
		Username: user,
		Password: key,
		Expire:   expire,
	}
	var item = keyring.Item{
		Key:         svcUser(user, service.Name),
		Data:        []byte(fmt.Sprintf("%d  %s", cred.Expire, key)),
		Label:       service.Label,
		Description: service.Description,
	}
	if err := s.Key[service].Set(item); err != nil {
		logPrint("error at setCredential/keyring.Set")
		return cred, fmt.Errorf("setCredential.Set:%s", err)
	}
	return cred, nil
}

// parseCredential function will take a string from the keyring service that has both the
// user password and the expire time and return a Credential structure with the parsed values.
func parseCredential(s keyring.Item) (Credential, error) {
	var resp Credential
	parsed := strings.Split(string(s.Data), "  ")
	if len(parsed) != 2 {
		return resp, fmt.Errorf("unable to parse secret, got len(%d) %v", len(parsed), parsed)
	}
	expire, err := strconv.ParseInt(parsed[0], 10, 64)
	if err != nil {
		logPrint("error at parseCredential/strconv.ParseInt")
		return resp, err
	}
	return Credential{
		Expire:   expire,
		Password: parsed[1],
	}, nil
}

// svcUser function will return the username used for the
// service we create for each token.
// We have to modify the user for backends that do not support a
// service name in their storage
func svcUser(user, service string) string {
	if isLinux() {
		return fmt.Sprintf("%s:%s", service, user)
	}
	return user
}

// isExpired method will take a expire time with a Credential receiver and
// determine if the Credential time is past the passed in expire.
func (c *Credential) isExpired() bool {
	return c.Expire < time.Now().Unix()
}
