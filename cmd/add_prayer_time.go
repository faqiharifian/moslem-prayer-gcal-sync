package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/auth"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/calendar"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/config"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/prayertime"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/server"
)

func AddPrayerTimeCmd(ctx context.Context, cfg config.Config) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Short: "Add prayer time from csv to google calendar",
		Use:   "add_prayer_time",
		Run: func(cmd *cobra.Command, args []string) {
			credFilePath := cmd.Flag("credentials").Value.String()
			cfg.LoadOauthConfig(credFilePath)

			prayerTimes, err := prayertime.FromCSV(filePath)
			if err != nil {
				fmt.Println("Failed to load csv file: ", err)
				return
			}
			prayerTimes.Filter()

			tokCh := make(chan *oauth2.Token)
			srv := server.New(cfg, prayerTimes, tokCh)
			authCli := auth.NewClient(cfg)

			go func() {
				server.Start(ctx, srv)
			}()

			go func() {
				authCli.Auth(calendar.AddPrayerTime, tokCh)
			}()

			tok := <-tokCh
			srv.Close()
			authCli.UpdateToken(tok)

			cal := calendar.NewClient(ctx, cfg.Oauth2, tok)
			err = cal.AddEvents(prayerTimes.ToEvents())

			if err != nil {
				fmt.Println("Error:", err)
			}

			fmt.Println("Successfully added prayer times")
		},
	}
	cmd.Flags().StringVarP(&filePath, "filepath", "f", "prayertime.csv", "Path to csv file")

	return cmd
}
