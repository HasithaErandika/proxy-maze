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

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackAttachment struct {
	Color  string       `json:"color"`
	Fields []slackField `json:"fields"`
	Footer string       `json:"footer"`
	Ts     int64        `json:"ts"`
}

type slackPayload struct {
	Username    string            `json:"username"`
	Text        string            `json:"text"`
	Attachments []slackAttachment `json:"attachments"`
}

func sendSlack(ig Integration, event string, a *alert.Alert) {
	color := "#FF0000"
	if event == "alert.resolved" {
		color = "#00AA00"
	}

	title := "ProxyMaze Alert: Proxy pool failure rate exceeded threshold"
	if event == "alert.resolved" {
		title = "ProxyMaze Alert: Proxy pool failure rate resolved"
	}

	payload := slackPayload{
		Username: ig.Username,
		Text:     title,
		Attachments: []slackAttachment{
			{
				Color: color,
				Fields: []slackField{
					{Title: "Alert ID", Value: a.AlertID, Short: true},
					{Title: "Failure Rate", Value: fmt.Sprintf("%.2f%%", a.FailureRate*100), Short: true},
					{Title: "Failed Proxies", Value: fmt.Sprintf("%d", a.FailedProxies), Short: true},
					{Title: "Threshold", Value: fmt.Sprintf("%.2f%%", a.Threshold*100), Short: true},
					{Title: "Failed IDs", Value: strings.Join(a.FailedProxyIDs, ", "), Short: false},
					{Title: "Fired At", Value: a.FiredAt.Format(time.RFC3339), Short: true},
				},
				Footer: "ProxyMaze'26 by Torch Labs",
				Ts:     time.Now().Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}
	webhook.DeliverWithRetry(ctx, client, ig.WebhookURL, body)
}
