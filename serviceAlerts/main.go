package main

import (
	"fmt"
	"github.com/codegold79/bael/emailAlerts"
	"github.com/codegold79/bael/gatherData"
	"github.com/codegold79/bael/userData"
)

func main() {
	baseUrl := "https://www.ltd.org/system-map/"
	routes, err := gatherData.GetAllRoutes()
	fmt.Println("Route list retrieved from db")

	allAlerts, err := gatherData.ScrapeSite(baseUrl, routes)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nwriting to file")
	gatherData.SaveAlertsToFile(allAlerts)
	fmt.Println("\nwriting to database")
	gatherData.UpdateDbAlerts(allAlerts)

	userKeys, err := userData.GetUserKeys()
	fmt.Println("User keys retrieved")
	if err != nil {
		fmt.Println(err)
	}
	// Retrieve service alert text and associated formatted route text from database, so
	// the info can be part of email alerts. Map keys are alert keys, and map values
	// are the service alert text and route info string (includes links to LTD site).
	alertsWithRoutes := gatherData.GetAlertsAndRoutesFromDb(routes, baseUrl)

	for _, uk := range userKeys {
		userData.UpdateUserAlerts(uk)
		fmt.Println("Outdated and unsubscribed users' alerts removed")

		userInfo, err := userData.GatherUserInfo(uk)

		if err != nil {
			fmt.Println(err)
		}

		// Gather new alert keys
		alertKeys, err := userData.GatherUserNewAlerts(uk, userInfo.Route_ids, userInfo.Stored_alert_keys)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Retrieved userData.")

		if len(alertKeys) > 0 {
			err = emailAlerts.SendEmail(userInfo.Email, alertKeys, alertsWithRoutes)
		} else {
			fmt.Println("No alerts to send.")
		}

		if err != nil {
			fmt.Println(err)
		}

		// Save keys in user data as they are no longer new.
		err = userData.SaveKeysInUserData(uk, alertKeys)

		if err != nil {
			fmt.Println(err)
		}
	}
}
