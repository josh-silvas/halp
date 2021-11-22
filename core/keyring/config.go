package keyring

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/josh5276/keyring"
	"github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"
)

const (
	configPath = ".config/gokeys"
	fileName   = "settings.ini"
)

// Settings type is the structure representation of
// the keyring ini profile held in the .config directory
type Settings struct {
	User         string
	JIRAInstance string
	JIRAUser     string
	Key          map[Service]keyring.Keyring
	pin          int
	File         *ini.File
	Source       string
	Test         bool
}

// CreateIfNotExist function will create the directory and file for the config
// if it does not already exist.  If this info already exist, it will do nothing.
func CreateIfNotExist(homeDir string) error {
	path := fmt.Sprintf("%s/%s", homeDir, configPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logPrint("directory does not exist, creating directory at ", path)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s", path, fileName)); os.IsNotExist(err) {
		logPrint("file does not exist, creating file ", fileName, path)
		file, err := os.Create(fmt.Sprintf("%s/%s", path, fileName))
		defer file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetConfig function takes a home directory path or none to use the user profile directory, and
// loads the ini file into a Settings structure and returns back the loaded config.
func GetConfig(homeDir string) (Settings, error) {
	if homeDir == "" {
		homeDir = userProfile.HomeDir
	}
	var (
		err      error
		settings Settings
	)
	settings.Source = fmt.Sprintf("%s/%s/%s", homeDir, configPath, fileName)

	settings.File, err = ini.InsensitiveLoad(settings.Source)
	if err != nil {
		logPrint("error at GetConfig/ini.InsensitiveLoad")
		return settings, err
	}
	if err := settings.loadBaseSection(settings.File); err != nil {
		return settings, err
	}

	if err := settings.File.SaveTo(settings.Source); err != nil {
		return settings, err
	}
	return settings, nil
}

func (s *Settings) loadBaseSection(cfg *ini.File) error {
	// Pull the base section locations
	sec, err := cfg.GetSection("")
	if err != nil {
		return err
	}

	// Set or prompt for the jira_token variable
	key, err := sec.GetKey("name")
	if err != nil {
		key, err = sec.NewKey("name", prompt("Please enter your name"))
		if err != nil {
			return err
		}
		key.Comment = "Full name"
	}
	s.User = key.String()

	// Set or prompt for the jira_token variable
	jiraInstance, err := sec.GetKey("jira_instance")
	if err != nil {
		jiraInstance, err = sec.NewKey("jira_instance", prompt("Enter the jira instance name (<company_name>.atlassian.net)"))
		if err != nil {
			return err
		}
		jiraInstance.Comment = "Jira Instance Name"
	}
	s.JIRAInstance = jiraInstance.String()

	// Set or prompt for the jira_token variable
	jiraUser, err := sec.GetKey("jira_username")
	if err != nil {
		jiraUser, err = sec.NewKey("jira_username", prompt("Enter the jira username (<username>@example.com)"))
		if err != nil {
			return err
		}
		jiraUser.Comment = "Jira User Name"
	}
	s.JIRAUser = jiraUser.String()

	// If we are using a supported keyring backend, then we don't need to set
	// a pin.
	for _, backend := range keyring.AvailableBackends() {
		switch backend {
		case keyring.KeychainBackend:
			return nil
		case keyring.WinCredBackend:
			return nil
		case keyring.SecretServiceBackend:
			return nil
		}
	}
	return s.getPin(sec)
}

func (s *Settings) getPin(sec *ini.Section) error {
	// Set or prompt for the sso_username variable
	pin, err := sec.GetKey("file_pin")
	if err != nil {
		pinInt := promptInt("Please enter a keychain 6 digit pin")
		pin, err = sec.NewKey("file_pin", strconv.Itoa(pinInt))
		if err != nil {
			return err
		}
		pin.Comment = "pin used to unlock file-based keyrings"
		s.pin = pinInt
		return nil
	}
	s.pin = pin.MustInt()
	return nil
}

// prompt is a simple helper function to prompt for missing
// config data.
func prompt(text string) string {
	resp, err := ui.Ask(text, &input.Options{Required: true, HideOrder: true})
	if err != nil {
		logrus.Fatal(err)
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	if resp == "" {
		return prompt(text)
	}
	return resp
}

// promptInt is a simple helper function to prompt for missing
// config data.
func promptInt(text string) (pin int) {
	resp, err := ui.Ask(text, &input.Options{Required: true, HideOrder: true})
	if err != nil {
		logrus.Fatal(err)
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	if resp == "" || len(resp) != 6 {
		return promptInt(text)
	}
	if pin, err = strconv.Atoi(resp); err != nil {
		return promptInt(text)
	}
	return
}

// promptSignature is function to pass into the keyring package
// that matches the signature required for the FilePrompt, but providing the
// credential's back into the keyring package.
func promptSignature(_ string) (string, error) {
	cfg, err := GetConfig("")
	if err != nil {
		return "", err
	}
	return strconv.Itoa(cfg.pin), nil
}
