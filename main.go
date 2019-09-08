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

type alerts []alert

type route struct {
	routeID   string
	routePath string
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

func scrapeSite(baseUrl string) (alerts, error) {
	var allAlerts alerts
	var routes []route

	routes, err := getRoutes()
	fmt.Println("Routes retrieved from db")

	if err != nil {
		return nil, err
	}

	url := baseUrl

	for _, v := range routes {
		url = baseUrl + v.routePath
		resp, err := http.Get(url)
		fmt.Println("Retrieved route from " + url)

		if err != nil {
			return allAlerts, err
		}

		someAlerts, err := parseHtml(resp, v.routeID)
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

func getRoutes() ([]route, error) {
	var routes []route
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
			routes = append(routes, route{rID.(string), rPath.(string)})
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

func parseHtml(page *http.Response, routeID string) (alerts, error) {
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
		txt := s.Find("div").Text()
		txt = strings.TrimSpace(txt)

		// Remove excessive tabs
		sli := strings.Split(txt, "\t")
		var sli2 []string
		for _, v := range sli {
			if v != "" {
				sli2 = append(sli2, v)
			}
		}
		txt = strings.Join(sli2, " ")

		// Remove excessive line breaks
		sli = strings.Split(txt, "\n")
		sli2 = []string{}
		for _, v := range sli {
			if v != " " {
				sli2 = append(sli2, v)
			}
		}
		txt = strings.Join(sli2, "\n")

		a = alert{txt, []string{routeID}}
		someAlerts = append(someAlerts, a)
	})

	return someAlerts, err
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
