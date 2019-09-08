package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"google.golang.org/api/iterator"
	"net/http"
	"os"
	"strings"
)

type alert struct {
	text     string
	routeIDs []string
}

func main() {
	var allAlerts alerts

	baseUrl := "https://www.ltd.org/system-map/"
	allAlerts, err := scrapeSite(baseUrl)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(allAlerts)
}

type alerts []alert

func scrapeSite(baseUrl string) (alerts, error) {
	var routes [][2]string

	routes, err := getRoutes()

	if err != nil {
		return nil, err
	}

	url := baseUrl

	var allAlerts alerts

	for _, v := range routes {
		url += v[1]
		resp, err := http.Get(url)

		if err != nil {
			return allAlerts, err
		}

		someAlerts, err := parseHtml(resp, v[0])
		err = someAlerts.addToAlerts(&allAlerts)

		if err != nil {
			return allAlerts, err
		}
	}

	return allAlerts, nil
}

func (someAlerts alerts) addToAlerts(allAlerts *alerts) error {
	for _, v := range someAlerts {
		*allAlerts = append(*allAlerts, v)
	}
	return nil
}

func getRoutes() ([][2]string, error) {
	var routes [][2]string
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		return routes, err
	}

	defer client.Close()

	iter := client.Collection("routes").Documents(ctx)

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		rID := doc.Data()["route_id"]
		rPath := doc.Data()["route_path"]

		if rID != nil && rPath != nil {
			routes = append(routes, [2]string{rID.(string), rPath.(string)})
		}
	}

	return routes, nil
}

func saveAlertsToDb(alerts map[int]string) error {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		fmt.Printf("firestore.NewClient: %v", err)
	}

	defer client.Close()

	// Grab all the documents that don't have outdated_at.
	iter := client.Collection("alerts").Where("outdated_at", "==", nil).Documents(ctx)

	for {
		// Go through the docs and see if the body is in the alerts map.
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		for k, v := range alerts {
			if v == doc.Data()["body"] {
				// If there is a matching entry in the alerts map, remove it.
				delete(alerts, k)
			} else {
				// If there is no match in alerts map, set outdated_at date.
				_, err := client.Collection("alerts").Doc(doc.Ref.ID).Set(ctx, map[string]interface{}{
					"outdated_at": firestore.ServerTimestamp,
				}, firestore.MergeAll)

				if err != nil {
					return err
				}
			}
		}
	}

	// Go through the remaining items in the alerts map. These should all be new alerts
	// save each alert in map as a new document.
	// type Alert struct {
	// 	Body        string    `firestore:"body,omitempty"`
	// 	Routes      []string  `firestore:"routes,omitempty"`
	// 	Outdated_at time.Time `firestore:"outdated_at,omitempty"`
	// }
	// var routes []string
	// for k, v := range alerts {
	// 	routes = getRoutes(alerts["body"])
	// }
	return nil
}

func parseHtml(page *http.Response, route string) (alerts, error) {
	var a alert
	var someAlerts alerts

	defer page.Body.Close()

	if page.StatusCode != 200 {
		return nil, fmt.Errorf("webpage didn't load")
	}

	doc, err := goquery.NewDocumentFromReader(page.Body)

	if err != nil {
		return nil, err
	}

	doc.Find(".alert").Each(func(i int, s *goquery.Selection) {
		a = alert{strings.TrimSpace(s.Find("div").Text()), []string{route}}
		someAlerts = append(someAlerts, a)
	})

	return someAlerts, err
}

// func parseHtml(content *http.Response) (alert, error) {
// 	defer content.Body.Close()
// 	htmlTokens := html.NewTokenizer(content.Body)
// 	var alert alert

// 	isAlertUl := false
// 	liOrder := 0
// 	i := 0

// 	for {
// 		tt := htmlTokens.Next()

// 		switch tt {
// 		case html.ErrorToken:
// 			return alert, htmlTokens.Err()
// 		case html.StartTagToken:
// 			// The start tag tokens of interest are
// 			// (1) ul that has the "alert_list" class. This indicates the start of collectListItems.
// 			// (2) First layer of li items. liOrder = 1
// 			// (3) Nested li items should be indicated as such, but need to be stored as part of the first layer.
// 			t := htmlTokens.Token()

// 			// (3) This is a ul inside the "alert_list" ul. Increment the li level to indicate
// 			// we want to keep adding to the most outside level of li data.
// 			if isAlertUl == true && t.Data == "ul" {
// 				liOrder++
// 			}

// 			// (1) alert_list ul has been found. Start saving data.
// 			if isAlertUl == false && len(t.Attr) > 0 && strings.Contains(t.Attr[0].Val, "alert_list") {
// 				isAlertUl = true
// 			}

// 			// (2) This is a new li in the first level. Start another element in the map.
// 			if isAlertUl == true && t.Data == "li" && liOrder == 0 {
// 				// Start new list item
// 				i++
// 			}

// 		case html.EndTagToken:
// 			t := htmlTokens.Token()

// 			// We have encountered the closing ul (or will be closing the ul) for our alert_list. We're done.
// 			if isAlertUl == true && liOrder == 0 && t.Data == "ul" {
// 				// We have the info we need. We can exit now.
// 				fmt.Println("Data collection complete")
// 				return alerts, nil
// 			}

// 			// Close the ul we found while in a more nested level.
// 			if isAlertUl == true && liOrder > 0 && t.Data == "ul" {
// 				liOrder--
// 			}
// 		case html.TextToken:
// 			if isAlertUl == true {
// 				t := htmlTokens.Token()
// 				text := strings.TrimSpace(t.Data)

// 				if alerts[i] != "" {
// 					alerts[i] = alerts[i] + " " + text
// 				} else {
// 					alerts[i] = text
// 				}
// 			}
// 		}
// 	}
// }

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
