package keyring

import (
	"fmt"
	"os"
	"os/user"
	"testing"
	"time"

	"github.com/josh5276/keyring"

	"github.com/sirupsen/logrus"
)

var (
	testCfg  Settings
	testCred = Credential{
		Username: "test_user",
		Password: "t0t@s_s3creT",
		Expire:   time.Unix(time.Now().Unix(), 0).Add(10 * time.Minute).Unix(),
	}
	testSvc = Service{
		Name:        "com.go.test.service",
		Label:       "Test Service",
		Description: "Test Description",
	}
)

func TestMain(m *testing.M) {
	logrus.Info("Setting up test keyring service...")
	var err error
	if testCfg, err = testNew(); err != nil {
		logrus.Fatalf("halp.keyring.New:%s", err)
	}
	os.Exit(m.Run())
}

func testNew() (s Settings, err error) {
	userProfile, err = user.Current()
	if err != nil {
		return s, fmt.Errorf("GetConfig:%s", err)
	}
	if err = CreateIfNotExist(userProfile.HomeDir); err != nil {
		return s, fmt.Errorf("user.Current():%s", err)
	}
	cfg, err := GetConfig("")
	if err != nil {
		return s, fmt.Errorf("GetConfig:%s", err)
	}
	cfg.Test = true

	// If we are at debug level in logrus, set debug in keyring
	keyring.Debug = logrus.GetLevel() == logrus.DebugLevel

	cfg.Key = make(map[Service]keyring.Keyring, 0)
	for _, svc := range []Service{testSvc} {
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
