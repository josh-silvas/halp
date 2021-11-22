package keyring

// Service type describes a credential that needs to be stored
// in the GoKeys keyring.
type Service struct {
	Name        string
	Label       string
	Description string
}

// Define new keyring data that needs to be install into the keychain here
var (
	// SvcTempo is an exportable type that describes the keyring data for Tempo
	SvcTempo = Service{
		Name:        "com.keyring.go.tempo",
		Label:       "Tempo Token",
		Description: "Tempo API Token",
	}

	// SvcJIRA is an exportable type that describes the keyring data for JIRA
	SvcJIRA = Service{
		Name:        "com.keyring.go.jira",
		Label:       "Jira Token",
		Description: "Jira API Token",
	}

	// SvcAll is an exportable type that describes the keyring data for
	// that can be used when access all Gokeys in a keychain
	SvcAll = Service{
		Name: "com.keyring.go.*",
	}

	// svcSlice declaration is used for functions that match the SvcAll type.
	svcSlice = []Service{
		SvcTempo,
		SvcJIRA,
	}
)
