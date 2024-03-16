package main

import (
	"context"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/cmd"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/config"
	"github.com/spf13/cobra"
)

func main() {
	if err := config.Load(); err != nil {
		panic(err)
	}
	cfg := config.Get()

	c := &cobra.Command{CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true}}
	c.PersistentFlags().StringP("credentials", "c", "credentials.json", "Credential json file location")
	c.AddCommand(cmd.AddPrayerTimeCmd(context.Background(), cfg))
	c.AddCommand(cmd.DeletePrayerTimeCmd(context.Background(), cfg))

	c.Execute()

	// ---------------
	// google.
	// cli, err := google.DefaultClient(context.Background(), calendar.CalendarEventsScope)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// credentials, err := google.FindDefaultCredentials(context.Background(), calendar.CalendarEventsScope)
	// fmt.Println(err)
	// fmt.Println(string(credentials.JSON))

	// token, err := credentials.TokenSource.Token()
	// // token.
	// fmt.Println(err)
	// fmt.Println(token.AccessToken)

	// srv, err := calendar.NewService(context.Background(), option.WithCredentials(credentials))
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve Calendar client: %v", err)
	// 	return
	// }
	// tm, _ := time.Parse(constant.DateFormatLayout, "01-03-2024")
	// events, err := srv.Events.List("primary").ShowDeleted(false).
	// 	SingleEvents(true).TimeMin(tm.Format(time.RFC3339)).MaxResults(10).OrderBy("startTime").Do()

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("event len: ", len(events.Items))
	// for _, event := range events.Items {
	// 	fmt.Println(event.Summary)
	// }
	// cal := mycal.NewClient(context.Background(), nil, token)
	// ------------

	// srv := server.New(cfg, &prayertime.PrayerTime{}, make(chan *oauth2.Token))
	// server.Start(context.Background(), srv)
}
