package main

import (
	"fmt"
	"github.com/apognu/gocal"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const icsDir = "rooster.ics"

var url = goDotEnvVariable("FEED_URL")

var lastUpdate = getFileDate()
var events = getRoster()

func main() {
	updateFile(icsDir, url)

	// Start api
	handleRequests()
}

func handleRequests() {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Gets all roster items for clasCode, if any.
	router.GET("/rooster/:classCode", getClassRoster())

	// Get all classes in current feed file. For now, only IC_INF* classes are available in current feed.
	router.GET("/classes", getClasses())

	router.Run()
}

func getClasses() func(c *gin.Context) {
	return func(c *gin.Context) {

		checkUpdate()
		roster := events

		var classCodes []string

		for _, e := range roster {
			var classes []string

			for _, attendee := range e.Attendees {
				if strings.Contains(attendee.Cn, "_") {
					classes = append(classes, attendee.Cn)
				} else {
				}
			}
			classCodes = append(classCodes, classes...)
		}

		// Filtering out duplicates after adding all is faster (180ms vs 7ms API response)
		c.IndentedJSON(http.StatusOK, unique(classCodes))
	}
}

func getRoster() []gocal.Event {

	// Timezone name in ics: Timezone name in go
	var tzMapping = map[string]string{
		"W. Europe Standard Time": "Europe/Amsterdam",
	}

	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		if tzid, ok := tzMapping[s]; ok {
			return time.LoadLocation(tzid)
		}
		return nil, fmt.Errorf("")
	})

	ics, _ := os.Open(icsDir)

	defer ics.Close()

	c := gocal.NewParser(ics)
	c.Parse()

	// Return all parsed events.
	return c.Events
}

func updateFile(dir string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(dir)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func getClassRoster() func(c *gin.Context) {
	return func(c *gin.Context) {

		checkUpdate()
		classCode := c.Param("classCode")
		roster := events

		var rosterItems []RosterItem

		for _, e := range roster {
			//fmt.Printf("%s | %s | %s |\n", e.Summary, e.Start, e.End)
			var teachers []string
			var classes []string

			for _, attendee := range e.Attendees {
				if strings.Contains(attendee.Cn, "_") {
					classes = append(classes, attendee.Cn)
				} else {
					teachers = append(teachers, attendee.Cn)
				}
			}
			if len(teachers) > 20 {
				teachers = nil
			}
			if len(classes) > 20 {
				classes = nil
			}
			if itemExists(classes, classCode) {
				var item = RosterItem{e.Summary, e.Location, teachers, classes, e.Start.String(), e.End.String(), e.Comment}
				rosterItems = append(rosterItems, item)
			}
		}

		c.IndentedJSON(http.StatusOK, rosterItems)
	}
}

func itemExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

func checkUpdate() {
	if isMoreThanTwoHoursAgo(lastUpdate) {
		fmt.Println("file updated.")
		updateFile(icsDir, url)
		events = getRoster()
		lastUpdate = getFileDate()
	}

}

func isMoreThanTwoHoursAgo(t time.Time) bool {
	return time.Now().Sub(t) > 2*time.Hour
}

func getFileDate() time.Time {
	file, err := os.Stat(icsDir)

	if err != nil {
		fmt.Println(err)
	}
	modifiedTime := file.ModTime()
	return modifiedTime
}

func unique[T comparable](s []T) []T {
	inResult := make(map[T]bool)
	var result []T
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file, please make sure it exists in the same dir and contains FEED_URL!")
	}

	return os.Getenv(key)
}

type RosterItem struct {
	Subject   string   `json:"subject"`
	Location  string   `json:"location"`
	Teachers  []string `json:"teachers"`
	Classes   []string `json:"classes"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
	Comment   string   `json:"comment"`
}
