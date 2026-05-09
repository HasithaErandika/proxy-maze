package integration

type Integration struct {
	Type       string   `json:"type"`
	WebhookURL string   `json:"webhook_url"`
	Username   string   `json:"username"`
	Events     []string `json:"events"`
}

type IntegrationReq struct {
	Type       string   `json:"type"`
	WebhookURL string   `json:"webhook_url"`
	Username   string   `json:"username"`
	Events     []string `json:"events"`
}
