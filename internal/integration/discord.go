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

type discordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type discordFooter struct {
	Text string `json:"text"`
}

type discordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Color       int            `json:"color"`
	Fields      []discordField `json:"fields"`
	Footer      discordFooter  `json:"footer"`
}

type discordPayload struct {
	Username string         `json:"username"`
	Embeds   []discordEmbed `json:"embeds"`
}

func sendDiscord(ig Integration, event string, a *alert.Alert) {
	color := 16711680 
	title := "Proxy Pool Alert Fired"
	desc := "Proxy pool failure rate has exceeded the threshold."

	if event == "alert.resolved" {
		color = 43520 
		title = "Proxy Pool Alert Resolved"
		desc = "Proxy pool failure rate is back below the threshold."
	}

	valFailedIDs := strings.Join(a.FailedProxyIDs, ", ")
	if valFailedIDs == "" {
		valFailedIDs = "None"
	}

	payload := discordPayload{
		Username: ig.Username,
		Embeds: []discordEmbed{
			{
				Title:       title,
				Description: desc,
				Color:       color,
				Fields: []discordField{
					{Name: "Alert ID", Value: a.AlertID, Inline: true},
					{Name: "Failure Rate", Value: fmt.Sprintf("%.2f%%", a.FailureRate*100), Inline: true},
					{Name: "Failed Proxies", Value: fmt.Sprintf("%d", a.FailedProxies), Inline: true},
					{Name: "Threshold", Value: fmt.Sprintf("%.2f%%", a.Threshold*100), Inline: true},
					{Name: "Failed IDs", Value: valFailedIDs, Inline: false},
				},
				Footer: discordFooter{
					Text: "ProxyMaze'26 • Torch Labs",
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
