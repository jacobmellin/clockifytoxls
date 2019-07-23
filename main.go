package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type timeEntry struct {
	ID           string      `json:"id"`
	Description  string      `json:"description"`
	TagIds       interface{} `json:"tagIds"`
	UserID       string      `json:"userId"`
	Billable     bool        `json:"billable"`
	TaskID       interface{} `json:"taskId"`
	ProjectID    interface{} `json:"projectId"`
	TimeInterval struct {
		Start    time.Time `json:"start"`
		End      time.Time `json:"end"`
		Duration string    `json:"duration"`
	} `json:"timeInterval"`
	WorkspaceID string `json:"workspaceId"`
	IsLocked    bool   `json:"isLocked"`
}

type config struct {
	WorkspaceID string `json:"workspaceId"`
	UserID      string `json:"userId"`
	APIKey      string `json:"apiKey"`
}

func getConfig() (config, error) {
	config := config{}
	file, err := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func main() {
	config, err := getConfig()

	if err != nil {
		log.Fatal(err)
		return
	}

	apiRoot := "https://clockify.me/api/v1/"
	timeEntriesRequestString := "workspaces/%s/user/%s/time-entries?start=%s&end=%s"

	var start string

	flag.StringVar(&start, "s", "", "A start date represented in the form YYYY-MM-DD.")
	flag.Parse()

	var startDateTime time.Time

	if start != "" {
		var err error
		startDateTime, err = time.ParseInLocation("2006-01-02", start, time.UTC)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		startDateTime = time.Date(2019, time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	endDateTime := time.Now().UTC()

	url := apiRoot + fmt.Sprintf(timeEntriesRequestString, config.WorkspaceID, config.UserID, startDateTime.Format(time.RFC3339), endDateTime.Format(time.RFC3339))

	fmt.Println("Start date:", startDateTime)
	fmt.Println("End date:", endDateTime)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	req.Header.Set("X-Api-Key", config.APIKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()

	t := make([]timeEntry, 0)

	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		log.Println(err)
	}

	sort.Slice(t, func(i, j int) bool {
		return t[i].TimeInterval.Start.Before(t[j].TimeInterval.End)
	})

	if err != nil {
		fmt.Println(err)
	}

	f := excelize.NewFile()
	// dateStyle, err := f.NewStyle(`{"number_format":22 }`)

	var currentMonth string
	var currentSheetLine int
	var sheet int

	// for i := len(t)/2 - 1; i >= 0; i-- {
	// 	opp := len(t) - 1 - i
	// 	t[i], t[opp] = t[opp], t[i]
	// }

	for _, e := range t {

		newMonth := e.TimeInterval.Start.Format("Jan")

		if newMonth != currentMonth {
			fmt.Println("Writing month " + newMonth)

			currentMonth = newMonth
			sheet = f.NewSheet(currentMonth)
			f.SetActiveSheet(sheet)
			f.DeleteSheet("Sheet1")

			f.SetCellValue(currentMonth, "A1", "Start")
			f.SetCellValue(currentMonth, "B1", "End")
			f.SetCellValue(currentMonth, "C1", "Duration")
			f.SetCellValue(currentMonth, "D1", "Task")
			f.SetCellValue(currentMonth, "E1", "Description")

			currentSheetLine = 0
		}

		cN, err := excelize.CoordinatesToCellName(1, currentSheetLine+2)

		if err != nil {
			fmt.Println(err)
			return
		}

		f.SetCellValue(currentMonth, cN, e.TimeInterval.Start.Add(2*time.Hour))
		// f.SetCellStyle(currentMonth, cN, cN, dateStyle)

		cN, err = excelize.CoordinatesToCellName(2, currentSheetLine+2)
		f.SetCellValue(currentMonth, cN, e.TimeInterval.End.Add(2*time.Hour))
		cN, err = excelize.CoordinatesToCellName(3, currentSheetLine+2)
		f.SetCellValue(currentMonth, cN, e.TimeInterval.End.Sub(e.TimeInterval.Start))
		descr := strings.Split(e.Description, ":")
		cN, err = excelize.CoordinatesToCellName(4, currentSheetLine+2)
		f.SetCellValue(currentMonth, cN, strings.Trim(descr[0], " "))
		cN, err = excelize.CoordinatesToCellName(5, currentSheetLine+2)
		f.SetCellValue(currentMonth, cN, strings.Trim(descr[1], " "))

		currentSheetLine++
	}

	err = f.SaveAs("./timesheet_" + time.Now().Format("2006-01-02") + ".xlsx")

	if err != nil {
		fmt.Println(err)
		return
	}
}
