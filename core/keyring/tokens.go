package keyring

import (
	"fmt"
	"time"
)

const tokenExpire = 720 * time.Hour

// TempoToken will take a token value from Tempo
func (s *Settings) TempoToken() (key Credential, err error) {
	logPrint("getting JIRA information...")
	if key, err = s.getCredential(s.User, SvcTempo); err == nil {
		return
	}

	tempoToken, err := ui.Ask("Please enter your Tempo Authentication Token", options)
	if err != nil {
		err = fmt.Errorf("Tempo:ui.Ask:%s", err)
		logPrint("error at Tempo/ui.Ask")
		return
	}
	expire := time.Unix(time.Now().Unix(), 0).Add(tokenExpire).Unix()

	return s.setCredential(s.User, tempoToken, SvcTempo, expire)
}

// JIRAToken will take a token value from JIRA
func (s *Settings) JIRAToken() (key Credential, err error) {
	logPrint("getting JIRA information...")
	if key, err = s.getCredential(s.User, SvcJIRA); err == nil {
		return
	}

	jiraToken, err := ui.Ask("Please enter your JIRA Authentication Token", options)
	if err != nil {
		err = fmt.Errorf("JIRA:ui.Ask:%s", err)
		logPrint("error at JIRA/ui.Ask")
		return
	}
	expire := time.Unix(time.Now().Unix(), 0).Add(tokenExpire).Unix()

	return s.setCredential(s.User, jiraToken, SvcJIRA, expire)
}
