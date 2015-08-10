package jira

type Errors struct {
	ErrorMessages []string `json:"errorMessages"`
}

type Session struct {
	Session struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"session"`
	LoginInfo struct {
		LoginCount        int    `json:"loginCount"`
		PreviousLoginTime string `json:"previousLoginTime"`
	} `json:"loginInfo"`
}

type Ticket struct {
	ID   string `json:"id"`
	Self string `json:"self"`
	Key  string `json:"key"`
}
