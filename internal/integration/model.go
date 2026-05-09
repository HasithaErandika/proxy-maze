package integration

// Integration represents a registered external integration.
type Integration struct {
	Type       string   `json:"type"`
	WebhookURL string   `json:"webhook_url"`
	Username   string   `json:"username"`
	Events     []string `json:"events"`
}

// IntegrationReq represents the incoming payload for POST /integrations
type IntegrationReq struct {
	Type       string   `json:"type"`
	WebhookURL string   `json:"webhook_url"`
	Username   string   `json:"username"`
	Events     []string `json:"events"`
}
