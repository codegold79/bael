package gatherData

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"google.golang.org/api/iterator"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type alert struct {
	text     string
	routeIDs []string
}

type alerts []alert

func ScrapeSite(baseUrl string, routes map[string]string) (alerts, error) {
	var allAlerts alerts

	url := baseUrl

	for id, path := range routes {
		url = baseUrl + path
		resp, err := http.Get(url)
		fmt.Println("Retrieved service alert(s) from " + url)

		if err != nil {
			return allAlerts, err
		}

		someAlerts, err := parseHtml(resp, id)
		allAlerts, err = someAlerts.addToAlerts(allAlerts)

		if err != nil {
			return allAlerts, err
		}
	}

	return allAlerts, nil
}

func (someAlerts alerts) addToAlerts(allAlerts alerts) (alerts, error) {
	var i int
	for _, v := range someAlerts {
		i = findIndexOfDupeAlert(v.text, allAlerts)
		if i > -1 {
			// This alert is a duplicate, so add route to existing alert.
			allAlerts[i].routeIDs = append(allAlerts[i].routeIDs, v.routeIDs[0])
		} else {
			allAlerts = append(allAlerts, v)
		}
	}
	return allAlerts, nil
}

func findIndexOfDupeAlert(text string, allAlerts alerts) int {
	i := -1

	for j, v := range allAlerts {
		if text == v.text {
			i = j
		}
	}

	return i
}

func GetAllRoutes() (map[string]string, error) {
	routes := make(map[string]string)

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
			routes[rID.(string)] = rPath.(string)
		}
	}

	return routes, nil
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
		txt := strings.TrimSpace(s.Find("div").Text())

		// Remove excessive spaces
		space := regexp.MustCompile(`(\s)(\s)*`)
		txt = space.ReplaceAllString(txt, "$1")

		a = alert{txt, []string{routeID}}
		someAlerts = append(someAlerts, a)
	})

	return someAlerts, err
}

func SaveAlertsToFile(alerts alerts) error {
	f, err := os.Create("outputs/ltd-service-alerts.txt")

	if err != nil {
		return err
	}

	defer f.Close()

	for _, v := range alerts {
		f.WriteString(strings.Join(v.routeIDs, ", ") + " " + v.text + "\n")
	}

	return nil
}

func UpdateDbAlerts(ltdAlerts alerts) error {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		fmt.Printf("firestore.NewClient: %v", err)
	}

	defer client.Close()

	currentDbAlerts := GetCurrentServiceAlertTextsFromDb()
	dbDocOutdated := true

	for dbKey, dbAlertText := range currentDbAlerts {
		dbDocOutdated = true

		for i, la := range ltdAlerts {
			if la.text == dbAlertText {
				// If there is a matching entry in the ltdAlerts, mark it for deletion.
				// It's not needed because it's already in the database, and the
				// database entry is still up-to-date.
				ltdAlerts[i].text = "delete"
				dbDocOutdated = false
			}
		}

		// Now that we've gone through all the ltdAlerts, we know those without
		// matching ltdAlert entries are outdated. So, set it as such in the db.
		if dbDocOutdated {
			err = SetDocAsOutdated(dbKey)
		}
	}

	// Create another slice with just the new alerts.
	var newAlerts []alert
	for _, v := range ltdAlerts {
		if v.text != "delete" {
			newAlerts = append(newAlerts, v)
		}
	}

	// Go through save all the alerts in the slice to the database.
	for _, v := range newAlerts {
		_, _, err = client.Collection("alerts").Add(ctx, map[string]interface{}{
			"alert_text":  v.text,
			"outdated_at": nil,
			"route_ids":   v.routeIDs,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func SetDocAsOutdated(docID string) error {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		fmt.Printf("firestore.NewClient: %v", err)
	}

	defer client.Close()

	_, err = client.Collection("alerts").Doc(docID).Set(ctx, map[string]interface{}{
		"outdated_at": firestore.ServerTimestamp,
	}, firestore.MergeAll)

	if err != nil {
		return err
	}

	return nil
}

// Store all the non-outdated service alert texts to reduce db queries.
func GetCurrentServiceAlertTextsFromDb() map[string]string {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	alerts := make(map[string]string)

	if err != nil {
		fmt.Printf("firestore.NewClient: %v", err)
	}

	defer client.Close()

	iter := client.Collection("alerts").Where("outdated_at", "==", nil).Documents(ctx)

	for {
		// Go through the docs and see if the body is in the alerts map.
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		alerts[doc.Ref.ID] = doc.Data()["alert_text"].(string)
	}

	return alerts
}

// Retrieve service alerts and associated route info to be able to include data
// in the alert emails.
func GetAlertsAndRoutesFromDb(routes map[string]string, baseUrl string) map[string][2]string {
	allAlertsAndRoutes := make(map[string][2]string)

	// What will be returned is a map whose keys are the alert keys and whose values
	// are the the alert text and formatted route info (includes URLs to the LTD side).
	var alertTextRoute [2]string
	type AlertInfo struct {
		Alert_text string   `firestore:"alert_text"`
		Route_ids  []string `firestore:"route_ids"`
	}
	var alertInfo AlertInfo

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		fmt.Printf("firestore.NewClient: %v", err)
	}

	defer client.Close()

	iter := client.Collection("alerts").Where("outdated_at", "==", nil).Documents(ctx)

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		err = doc.DataTo(&alertInfo)

		routeStr, err := FormatRoutes(alertInfo.Route_ids, baseUrl)

		if err != nil {
			fmt.Printf("Route info could not be formatted: %v", err)
		}

		alertTextRoute = [2]string{alertInfo.Alert_text, routeStr}

		allAlertsAndRoutes[doc.Ref.ID] = alertTextRoute
	}

	return allAlertsAndRoutes
}

func FormatRoutes(routes []string, baseUrl string) (string, error) {
	var routesText string
	type RouteInfo struct {
		Name       string `firestore:"name"`
		Route_id   string `firestore:"route_id"`
		Route_path string `firestore:"route_path"`
	}
	var routeInfo RouteInfo
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		return "", err
	}

	for _, route := range routes {
		iter := client.Collection("routes").Where("route_id", "==", route).Documents(ctx)
		for {
			doc, err := iter.Next()

			if err == iterator.Done {
				break
			}

			err = doc.DataTo(&routeInfo)

			if err != nil {
				return routesText, err
			}

			routesText += fmt.Sprintf("%v (<a href='%v/%v'>Route %v</a>), ", routeInfo.Name, baseUrl, routeInfo.Route_path, routeInfo.Route_id)
		}
	}

	// Remove last comma and space.
	routesText = routesText[:len(routesText)-2]

	return routesText, nil
}
