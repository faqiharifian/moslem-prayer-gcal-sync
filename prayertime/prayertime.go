package prayertime

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/constant"
)

const (
	dateHeaderName     = "date.gregorian.date"
	dayHeaderName      = "date.gregorian.weekday.en"
	timezoneHeaderName = "meta.timezone"
	timeLayout         = "02-01-2006 15:04 (MST)"
)

var (
	days = []string{
		time.Sunday.String(),
		time.Monday.String(),
		time.Tuesday.String(),
		time.Wednesday.String(),
		time.Thursday.String(),
		time.Friday.String(),
		time.Saturday.String(),
	}

	validVisibilities = []string{
		"default",
		"private",
		"public",
	}
)

type PrayerTime struct {
	HeaderIdx       map[string]int
	SelectedTimings []string
	Records         [][]string
	Data            []PrayerTimeData
	SelectedDays    []string
	Visibility      string
}

type PrayerTimeData struct {
	Date    time.Time
	Day     string
	Timings map[string]time.Time
}

func FromCSV(path string) (*PrayerTime, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("the file is empty")
	}

	headers := map[string]int{}
	timingsKeys := []string{}

	data := []PrayerTimeData{}

	for i, row := range records {
		if i == 0 {
			for j, cell := range row {
				if strings.HasPrefix(cell, "timings.") {
					key := strings.TrimPrefix(cell, "timings.")
					timingsKeys = append(timingsKeys, key)
					headers[key] = j
				}
				if cell == dateHeaderName || cell == timezoneHeaderName || cell == dayHeaderName {
					headers[cell] = j
				}
			}
			if len(timingsKeys) == 0 || len(headers) < len(timingsKeys)+3 {
				return nil, fmt.Errorf("missing required columns")
			}
			continue
		}

		dateStr := row[headers[dateHeaderName]]
		date, err := time.Parse(constant.DateFormatLayout, dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date: %s, err: %w", dateStr, err)
		}

		tzStr := row[headers[timezoneHeaderName]]
		tz, err := time.LoadLocation(tzStr)
		if err != nil {
			return nil, err
		}

		dt := PrayerTimeData{
			Date:    date,
			Day:     row[headers[dayHeaderName]],
			Timings: map[string]time.Time{},
		}

		for _, key := range timingsKeys {
			timeStr := dateStr + " " + row[headers[key]]
			time, err := time.ParseInLocation(timeLayout, timeStr, tz)
			if err != nil {
				return nil, fmt.Errorf("invalid time: %s, err: %w", timeStr, err)
			}
			dt.Timings[key] = time
		}

		data = append(data, dt)
	}

	return &PrayerTime{
		HeaderIdx:       headers,
		Records:         records,
		Data:            data,
		SelectedTimings: timingsKeys,
		SelectedDays:    days,
	}, nil
}

func (p *PrayerTime) Filter() {
	fmt.Printf("You will add prayer times from %s to %s to your Google Calendar.\n", p.Data[0].Date.Format(constant.DateFormatLayout), p.Data[len(p.Data)-1].Date.Format(constant.DateFormatLayout))
	p.SelectedTimings = p.getSelections(p.SelectedTimings, "Please select prayer times to add:")

	fmt.Printf("These prayer times will be added: %s.\n", strings.Join(p.SelectedTimings, ", "))

	p.SelectedDays = p.getSelections(p.SelectedDays, "Please choose the days:")

	fmt.Printf("Prayer times will be added to these days: %s.\n", strings.Join(p.SelectedDays, ", "))

	msg := fmt.Sprintf("Please select visibility for google calendar (%s): ", strings.Join(validVisibilities, ", "))
	p.Visibility = p.getSelection(validVisibilities, msg)
	fmt.Println()
}

func (p *PrayerTime) ToEvents() []*calendar.Event {
	events := []*calendar.Event{}
	for _, data := range p.Data {
		if !slices.Contains(p.SelectedDays, data.Day) {
			continue
		}

		for _, timing := range p.SelectedTimings {
			prayerTime := data.Timings[timing]
			startTime := prayerTime.Format(time.RFC3339)
			endTime := prayerTime.Add(45 * time.Minute).Format(time.RFC3339)
			if data.Day == time.Friday.String() {
				startTime = prayerTime.Add(-30 * time.Minute).Format(time.RFC3339)
				endTime = prayerTime.Add(time.Hour).Format(time.RFC3339)
			}
			event := &calendar.Event{
				Summary:     fmt.Sprintf("%s Prayer", timing),
				Description: constant.CalendarWatermark,
				Start: &calendar.EventDateTime{
					DateTime: startTime,
				},
				End: &calendar.EventDateTime{
					DateTime: endTime,
				},
				Reminders: &calendar.EventReminders{
					Overrides: []*calendar.EventReminder{
						{
							Method:  "popup",
							Minutes: 5,
						},
					},
					UseDefault:      false,
					ForceSendFields: []string{"UseDefault"},
				},
				Visibility: p.Visibility,
				EventType:  "focusTime",
			}
			events = append(events, event)
		}
	}

	return events
}

func (p *PrayerTime) getSelection(options []string, msg string) string {
	for {
		fmt.Printf(msg)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		selection := strings.TrimSpace(scanner.Text())
		fmt.Println()

		if !slices.Contains(options, selection) {
			fmt.Println("Input is invalid, please try again.")
			continue
		}

		return selection
	}
}

func (p *PrayerTime) getSelections(options []string, msg string) []string {
	fmt.Println(msg)
	validOpts := make([]int, 0, len(options))
	for i, timing := range options {
		optNum := i + 1
		fmt.Printf("%d. %s\n", optNum, timing)
		validOpts = append(validOpts, optNum)
	}

	selections := []int{}
	for {
		selections = []int{}

		fmt.Printf("Use comma to select multiple items (for example: 1,2,4): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		selectionStr := scanner.Text()
		fmt.Println()

		selectionsStr := strings.Split(selectionStr, ",")
		if len(selectionsStr) == 0 {
			fmt.Println("You have not choose anything. Please try again.")
			continue
		}

		isValid := true
		for _, selection := range selectionsStr {
			selectionNum, _ := strconv.Atoi(strings.TrimSpace(selection))
			if !slices.Contains(validOpts, selectionNum) {
				fmt.Println("Selection is not valid. Please try again.")
				isValid = false
				break
			}
			selections = append(selections, selectionNum)
		}
		if !isValid {
			continue
		}
		break
	}

	sort.Ints(selections)

	selectedOpts := make([]string, 0, len(selections))
	for _, selection := range selections {
		selectedOpts = append(selectedOpts, options[selection-1])
	}
	return selectedOpts
}
