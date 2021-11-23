package atlassian

import "time"

type (
	// WorkLogResponse : structure to represent the response payload for a /workload requests.
	WorkLogResponse struct {
		Self     string `json:"self"`
		Metadata struct {
			Count  int    `json:"count"`
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Next   string `json:"next"`
		} `json:"metadata"`
		Results []Worklog `json:"results"`
	}

	// Worklog : structure for a worklog entry.
	Worklog struct {
		Self           string `json:"self"`
		TempoWorklogID int    `json:"tempoWorklogId"`
		JiraWorklogID  int    `json:"jiraWorklogId"`
		Issue          struct {
			Self string `json:"self"`
			Key  string `json:"key"`
			ID   int    `json:"id"`
		} `json:"issue"`
		TimeSpentSeconds int       `json:"timeSpentSeconds"`
		BillableSeconds  int       `json:"billableSeconds"`
		StartDate        string    `json:"startDate"`
		StartTime        string    `json:"startTime"`
		Description      string    `json:"description"`
		CreatedAt        time.Time `json:"createdAt"`
		UpdatedAt        time.Time `json:"updatedAt"`
		Author           struct {
			Self        string `json:"self"`
			AccountID   string `json:"accountId"`
			DisplayName string `json:"displayName"`
		} `json:"author"`
		Attributes struct {
			Self   string        `json:"self"`
			Values []interface{} `json:"values"`
		} `json:"attributes"`
	}

	// JIRAIssue : structure that reprosents the response payload of a JIRA issue.
	JIRAIssue struct {
		Expand string `json:"expand"`
		ID     string `json:"id"`
		Self   string `json:"self"`
		Key    string `json:"key"`
		Fields struct {
			Priority struct {
				Self    string `json:"self"`
				IconURL string `json:"iconUrl"`
				Name    string `json:"name"`
				ID      string `json:"id"`
			} `json:"priority"`
			Assignee interface{} `json:"assignee"`
			Status   struct {
				Self           string `json:"self"`
				Description    string `json:"description"`
				IconURL        string `json:"iconUrl"`
				Name           string `json:"name"`
				ID             string `json:"id"`
				StatusCategory struct {
					Self      string `json:"self"`
					ID        int    `json:"id"`
					Key       string `json:"key"`
					ColorName string `json:"colorName"`
					Name      string `json:"name"`
				} `json:"statusCategory"`
			} `json:"status"`
			Project struct {
				Self string `json:"self"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
			Created string `json:"created"`
			Updated string `json:"updated"`
			Summary string `json:"summary"`
		} `json:"fields"`
	}

	// IssueRequest : structure to create a new issue in Atlassian JIRA.
	IssueRequest struct {
		Fields IssueField `json:"fields"`
	}

	// IssueField : Nested field structure.
	IssueField struct {
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	}

	// IssueResponse : When a new issue is created, this will be the response payload.
	IssueResponse struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Self string `json:"self"`
	}
)
