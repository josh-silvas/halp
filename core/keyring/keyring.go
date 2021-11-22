package keyring

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/josh5276/keyring"
	"github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"
)

const (
	// KeyChainName declares a separate keychain as to separate from other keychains
	// that could possibly sync to iCloud or other devices. This keychain will not be able
	// to synchronize
	KeyChainName = "Go Keyring Internal"

	fileBackend = "~/.local/share/keyrings/"
	keychainCMD = "/usr/bin/security"
)

var (
	ui      = &input.UI{Writer: os.Stdout, Reader: os.Stdin}
	options = &input.Options{Required: true, Mask: true, HideOrder: true}

	userProfile *user.User
	logger      Logger
)

// Logger defines a signature type that should be used to pass in any
// logger types into the package.
type Logger func(v ...interface{})

// New function will initialize a logger type, gather profile information
// and setup the config directory if needed.
func New(logImport Logger) (s Settings, err error) {
	userProfile, err = user.Current()
	if err != nil {
		return s, err
	}
	if err = CreateIfNotExist(userProfile.HomeDir); err != nil {
		return s, err
	}
	logger = logImport
	cfg, err := GetConfig("")
	if err != nil {
		return s, err
	}

	// If we are at debug level in logrus, set debug in keyring
	keyring.Debug = logrus.GetLevel() == logrus.DebugLevel

	cfg.Key = make(map[Service]keyring.Keyring, 0)
	for _, svc := range svcSlice {
		cfg.Key[svc], err = keyring.Open(keyring.Config{
			AllowedBackends: []keyring.BackendType{
				keyring.KeychainBackend,
				keyring.WinCredBackend,
				keyring.FileBackend,
			},
			ServiceName: svc.Name,

			// Needed for default file fallback
			FileDir:          fileBackend,
			FilePasswordFunc: promptSignature,

			// MacOS default items
			KeychainName:                   KeyChainName,
			KeychainTrustApplication:       true,
			KeychainSynchronizable:         false,
			KeychainAccessibleWhenUnlocked: true,
		})
		if err != nil {
			logrus.Errorf("gokeys:Open:%s:%s", svc.Name, err)
		}
	}

	// Check if the new keychain is unlocked. If not
	// process the unlock command.
	return cfg, keychainUnlock(cfg)
}

// logPrint function uses the Logger method associated with the non exported value.
func logPrint(v ...interface{}) {
	if logger == nil {
		return
	}
	logger(v...)
}

func keychainUnlock(cfg Settings) error {
	if !isMacOS() {
		return nil
	}

	if !cfg.Test {
		// Attempt a pull for JIRA token to see if the keychain exist, or if we need
		// to create a new one.
		if _, err := cfg.TempoToken(); err != nil {
			return fmt.Errorf("keychainUnlock.cfg.TempoToken:%s", err)
		}
	}

	var keychainDB = fmt.Sprintf("%s.keychain-db", KeyChainName)

	out, err := exec.Command(keychainCMD, "show-keychain-info", keychainDB).CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychainUnlock.exec.Command:show-keychain-info failed %s:%s", KeyChainName, err)
	}

	// If there is no-timeout set on the keychain itself, this is what we want.
	// Exit here.
	if strings.Contains(fmt.Sprintf("%s", out), "no-timeout") {
		return nil
	}

	_, err = exec.Command(keychainCMD, "set-keychain-settings", keychainDB).CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychainUnlock.exec.Command:set-keychain-settings failed %s:%s", KeyChainName, err)
	}
	return nil
}

// Delete function will clean any associated keyrings for a given
// user.
func (s *Settings) Delete(service Service) error {
	if service != SvcAll {
		if err := s.Key[service].Remove(svcUser(s.User, service.Name)); err == nil {
			logrus.Infof("deleted %s key", service.Name)
		}
		return nil
	}

	for _, svc := range svcSlice {
		if err := s.Key[svc].Remove(svcUser(s.User, service.Name)); err == nil {
			logrus.Infof("deleted %s key", svc.Name)
		}
	}
	return nil
}

func isMacOS() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	for _, platform := range []string{"linux", "freebsd", "linux2"} {
		if platform == runtime.GOOS {
			return true
		}
	}
	return false
}
