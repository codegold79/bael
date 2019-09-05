package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
)

func main() {
	url := "https://www.ltd.org/service-alerts/"
	resp, err := http.Get(url)

	alerts := make(map[int]string)

	if err != nil {
		fmt.Print(err)
	}

	alerts, err = parseHtml(resp)

	if err != nil {
		fmt.Print(err)
	}

	err = saveAlertsToFile(alerts)
}

func parseHtml(content *http.Response) (map[int]string, error) {
	defer content.Body.Close()
	htmlTokens := html.NewTokenizer(content.Body)
	alerts := make(map[int]string)

	isAlertUl := false
	liOrder := 0
	i := 0

	for {
		tt := htmlTokens.Next()

		switch tt {
		case html.ErrorToken:
			return alerts, htmlTokens.Err()
		case html.StartTagToken:
			// The start tag tokens of interest are
			// (1) ul that has the "alert_list" class. This indicates the start of collectListItems.
			// (2) First layer of li items. liOrder = 1
			// (3) Nested li items should be indicated as such, but need to be stored as part of the first layer.
			t := htmlTokens.Token()

			// (3) This is a ul inside the "alert_list" ul. Increment the li level to indicate
			// we want to keep adding to the most outside level of li data.
			if isAlertUl == true && t.Data == "ul" {
				liOrder++
			}

			// (1) alert_list ul has been found. Start saving data.
			if isAlertUl == false && len(t.Attr) > 0 && strings.Contains(t.Attr[0].Val, "alert_list") {
				isAlertUl = true
			}

			// (2) This is a new li in the first level. Start another element in the map.
			if isAlertUl == true && t.Data == "li" && liOrder == 0 {
				// Start new list item
				i++
			}

		case html.EndTagToken:
			t := htmlTokens.Token()

			// We have encountered the closing ul (or will be closing the ul) for our alert_list. We're done.
			if isAlertUl == true && liOrder == 0 && t.Data == "ul" {
				// We have the info we need. We can exit now.
				fmt.Println("Done with the data ")
				return alerts, nil
			}

			// Close the ul we found while in a more nested level.
			if isAlertUl == true && liOrder > 0 && t.Data == "ul" {
				liOrder--
			}
		case html.TextToken:
			if isAlertUl == true {
				t := htmlTokens.Token()
				text := strings.TrimSpace(t.Data)

				if alerts[i] != "" {
					alerts[i] = alerts[i] + " " + text
				} else {
					alerts[i] = text
				}
			}
		}
	}
}

func saveAlertsToFile(content map[int]string) error {
	f, err := os.Create("outputs/ltd-service-alerts.txt")

	if err != nil {
		return err
	}

	defer f.Close()
	fmt.Println("\nwriting to file")

	for _, v := range content {
		f.WriteString(v + "\n")
	}

	return nil
}
