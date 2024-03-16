package calendar

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/constant"
)

const (
	AddPrayerTime    = "add-prayer-time"
	DeletePrayerTime = "delete-prayer-time"
)

type Client struct {
	token *oauth2.Token
	cli   *calendar.Service
}

func NewClient(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) *Client {
	cli := cfg.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(cli))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	return &Client{
		token: token,
		cli:   srv,
	}
}

func (c *Client) AddEvents(events []*calendar.Event) error {
	for _, event := range events {
		_, err := c.cli.Events.Insert("primary", event).Do()
		if err != nil {
			return fmt.Errorf("failed insert event for %s at %s, err: %w", event.Summary, event.Start.DateTime, err)
		}
		startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		fmt.Printf("Added: '%s' on %s\n", event.Summary, startTime.Format(constant.DateFormatLayout))
	}
	fmt.Println()

	return nil
}

func (c *Client) DeleteEvents(ctx context.Context, from, to time.Time) error {
	timeMin := from.Format(time.RFC3339)
	timeMax := to.Format(time.RFC3339)
	err := c.cli.Events.List("primary").
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		Pages(ctx, func(events *calendar.Events) error {
			for _, event := range events.Items {
				if !strings.Contains(event.Description, constant.CalendarWatermark) {
					continue
				}
				err := c.cli.Events.Delete("primary", event.Id).Do()
				if err != nil {
					return err
				}
				startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
				fmt.Printf("Deleted: '%s' on %s\n", event.Summary, startTime.Format(constant.DateFormatLayout))
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("failed getting events, err: %w", err)
	}
	fmt.Println()
	return nil
}
