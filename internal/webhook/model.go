package webhook

// Webhook represents a registered webhook receiver.
type Webhook struct {
	ID  string `json:"webhook_id"`
	URL string `json:"url"`
}

// Payload represents the JSON structure sent to webhook URLs.
type Payload struct {
	Event string      `json:"event"`
	Alert interface{} `json:"alert"`
}
