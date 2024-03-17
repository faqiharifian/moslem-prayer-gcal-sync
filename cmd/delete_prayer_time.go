package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/auth"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/calendar"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/config"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/constant"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/prayertime"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/server"
)

func DeletePrayerTimeCmd(ctx context.Context, cfg config.Config) *cobra.Command {
	var startDate, endDate string
	cmd := &cobra.Command{
		Short: "Delete prayer times from google calendar between start date and end date",
		Use:   "delete_prayer_time",
		Run: func(cmd *cobra.Command, args []string) {
			credFilePath := cmd.Flag("credentials").Value.String()
			cfg.LoadOauthConfig(credFilePath)

			startDate, err := time.ParseInLocation(constant.DateFormatLayout, startDate, time.Local)
			if err != nil {
				fmt.Println("Error: invalid start date, use format: DD-MM-YYYY (eg. 01-03-2024)")
			}
			endDate, err := time.ParseInLocation(constant.DateFormatLayout, endDate, time.Local)
			if err != nil {
				fmt.Println("Error: invalid end date, use format: DD-MM-YYYY (eg. 01-03-2024)")
			}

			fmt.Printf("You are going to delete ALL prayer time from %s to %s. Continue (y/N)? ", startDate.Format(constant.DateFormatLayout), endDate.Format(constant.DateFormatLayout))
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := scanner.Text()

			if !slices.Contains([]string{"y", "Y"}, answer) {
				return
			}
			fmt.Println()

			prayerTimes := &prayertime.PrayerTime{
				Data: []prayertime.PrayerTimeData{
					{Date: startDate},
					{Date: endDate},
				},
			}

			tokCh := make(chan *oauth2.Token)
			srv := server.New(cfg, prayerTimes, tokCh)
			authCli := auth.NewClient(cfg)

			go func() {
				server.Start(ctx, srv)
			}()

			go func() {
				authCli.Auth(calendar.DeletePrayerTime, tokCh)
			}()

			tok := <-tokCh
			srv.Close()
			authCli.UpdateToken(tok)

			cal := calendar.NewClient(ctx, cfg.Oauth2, tok)
			err = cal.DeleteEvents(ctx, startDate, endDate)
			if err != nil {
				fmt.Println("Error:", err)
			}

			fmt.Println("Successfully deleted prayer times")
		},
	}
	cmd.Flags().StringVar(&startDate, "start", "", "Start date to delete prayer time. Format: DD-MM-YYYY (eg. 01-03-2024)")
	cmd.Flags().StringVar(&endDate, "end", "", "End date to delete prayer time. Format: DD-MM-YYYY (eg. 31-03-2024)")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}
