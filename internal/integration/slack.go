package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackBlock struct {
	Type     string      `json:"type"`
	Text     *slackText  `json:"text,omitempty"`
	Fields   []slackText `json:"fields,omitempty"`
	Elements []slackText `json:"elements,omitempty"`
}

type slackPayload struct {
	Username string       `json:"username"`
	Blocks   []slackBlock `json:"blocks"`
}

func sendSlack(ig Integration, event string, a *alert.Alert) {
	title := "ProxyMaze Alert: Proxy pool failure rate exceeded threshold"
	if event == "alert.resolved" {
		title = "ProxyMaze Alert: Proxy pool failure rate resolved"
	}

	failedIDs := strings.Join(a.FailedProxyIDs, ", ")
	if failedIDs == "" {
		failedIDs = "None"
	}

	payload := slackPayload{
		Username: ig.Username,
		Blocks: []slackBlock{
			{
				Type: "header",
				Text: &slackText{Type: "plain_text", Text: title},
			},
			{
				Type: "section",
				Fields: []slackText{
					{Type: "mrkdwn", Text: "*Alert ID:*\n" + a.AlertID},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Failure Rate:*\n%.2f%%", a.FailureRate*100)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Failed Proxies:*\n%d", a.FailedProxies)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Threshold:*\n%.2f%%", a.Threshold*100)},
				},
			},
			{
				Type: "section",
				Text: &slackText{
					Type: "mrkdwn",
					Text: "*Failed IDs:*\n" + failedIDs,
				},
			},
			{
				Type: "context",
				Elements: []slackText{
					{Type: "mrkdwn", Text: "ProxyMaze'26 by Torch Labs"},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()

	client := &http.Client{}
	webhook.DeliverWithRetry(ctx, client, ig.WebhookURL, body)
}
