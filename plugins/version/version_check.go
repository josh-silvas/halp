package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/josh5276/halp/core/keyring"

	"github.com/Masterminds/semver"
	"github.com/go-ini/ini"
	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

const (
	checkInterval = 2
	versionAPI    = "https://api.github.com/repos/josh5276/halp/tags"
)

// Check function is executed from the halp caller
func Check(cfg keyring.Settings, version string) error {
	runningVer := SemVer(version)
	key, err := FromCfg(cfg)
	if err != nil {
		return fmt.Errorf("version check failed: %s", err)
	}
	storedVer, err := Parse(key.String())
	if err != nil {
		return fmt.Errorf("version check failed: %s", err)
	}
	key.SetValue(CfgVer{Version: runningVer, Timestamp: time.Now()}.String())
	if err = cfg.File.SaveTo(cfg.Source); err != nil {
		return fmt.Errorf("version check failed: %s", err)
	}

	// Here we are checking if the timestamp on the cached version is more than
	// 12 hours old. If it's not then we can just exit here.
	if storedVer.Timestamp.After(time.Now().Add(-checkInterval * time.Hour)) {
		return nil
	}

	apiVer, err := FromAPI()
	if err != nil {
		return fmt.Errorf("version check failed: %s", err)
	}
	if runningVer.LessThan(apiVer) {
		Notify(runningVer, apiVer)
	}
	return nil
}

// CfgVer represents the parsed type from the
// configuration file stored locally
type CfgVer struct {
	Version   *semver.Version
	Timestamp time.Time
}

// String method set onto the CfgVer type will convert the type into a
// value that is expected and consistent so that it can be parsed later.
func (c CfgVer) String() string {
	// As a standard, we are going to use time.RFC3339 as the timestamp storage format.
	return fmt.Sprintf("%s::%s", c.Version, c.Timestamp.Format(time.RFC3339))
}

// Parse will take a string value and attempt to parse is into a CfgVer type.
// If the string is set to its null value, then return an empty type and no error as
// this would be the accurate representation of the parsed item.
func Parse(c string) (CfgVer, error) {
	// Do not error if the string is empty, simply return
	// the empty value of the CfgVer type.
	if c == "" {
		return CfgVer{}, nil
	}
	arr := strings.Split(c, "::")
	if len(arr) != 2 {
		return CfgVer{}, fmt.Errorf("unable to parse version from cfg: %s", c)
	}
	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(arr[1]))
	if err != nil {
		return CfgVer{}, err
	}
	ver, err := semver.NewVersion(arr[0])
	if err != nil {
		return CfgVer{}, err
	}
	return CfgVer{Version: ver, Timestamp: ts}, nil
}

// FromCfg will take a gokeys.Settings type to open the config file and retrieve the
// cached version and timestamp within the config file. If there is no item in the file
// it will create the element and set the value to a string empty value, or ""
func FromCfg(cfg keyring.Settings) (*ini.Key, error) {
	sec, err := cfg.File.GetSection("halp")
	// If there is an error retrieving the section, it likely is not created yet.
	// attempt to create the section
	if err != nil {
		sec, err = cfg.File.NewSection("halp")
		if err != nil {
			return nil, err
		}

	}

	// Get the version section key. If it does not exist,
	// create the key and set the value to ""
	key, err := sec.GetKey("version")
	if err != nil {
		key, err = sec.NewKey("version", "v0.0.0::0001-01-01T00:00:00Z")
		if err != nil {
			return nil, err
		}
		key.Comment = "local cache of the halp version, really for the timestamp"
	}
	return key, nil
}

// APIResp type is the parsed value returned from the API
// that stores halp's current operating version
type APIResp struct {
	Name       string `json:"name"`
	ZipBallURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

// FromAPI function will get the standard version from the migration API
func FromAPI() (*semver.Version, error) {
	var res = make([]APIResp, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	// Build the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionAPI, nil)
	if err != nil {
		logrus.Errorf("version:Fetch:NewRequest:%s", err)
		return nil, err
	}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("version:FromAPI:Do:%s", err)
		return nil, err
	}
	if r != nil {
		defer func() {
			if defErr := r.Body.Close(); defErr != nil {
				err = fmt.Errorf("%s/%s", err, defErr)
			}
		}()
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("version:FromAPI:Fetch:%s", r.Status)
	}

	if err = json.NewDecoder(r.Body).Decode(&res); err != nil {
		return nil, err
	}
	latestVer, err := semver.NewVersion("0.0.0")
	if err != nil {
		return nil, err
	}
	for _, tag := range res {
		v, err := semver.NewVersion(tag.Name)
		if err != nil {
			return nil, err
		}
		if v.GreaterThan(latestVer) {
			latestVer = v
		}
	}

	return latestVer, nil
}

// SemVer is a helper to convert goreleaser/git tags with
// semver tags.
func SemVer(s string) *semver.Version {
	if s == "" {
		s = "0.0.0"
	}
	ver, err := semver.NewVersion(strings.TrimPrefix(s, "v"))
	if err != nil {
		logrus.Errorf("Version check failed: %s:%s", err, s)
	}
	return ver
}

// Notify is used to print info to terminal if the user needs
// to be notified of a new or different running version
func Notify(running, current *semver.Version) {
	color.LightYellow.Printf("Upgrade available (%s running, %s available). Install with:\n", running, current)
	switch runtime.GOOS {
	case "linux":
		color.LightYellow.Printf("   >> curl -O https://<package_url>/halp_64-bit.deb "+
			"&& sudo dpkg -i halp_64-bit.deb\n", current)
		color.Yellow.Printf("You will be notified in %d hours if you have not upgraded.\n", checkInterval)
	case "darwin":
		color.LightYellow.Printf("   >> brew update && brew upgrade halp\n")
		color.Yellow.Printf("You will be notified in %d hours if you have not upgraded.\n", checkInterval)
	default:
		color.Yellow.Print("Unknown OS, check https://github.com/josh5276/halp for install options\n")
	}
}
