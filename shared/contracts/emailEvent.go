package contracts

// EmailEvent defines the structure for any email message event
type EmailEvent struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
