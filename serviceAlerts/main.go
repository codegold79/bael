package main

import (
	"fmt"
	"github.com/codegold79/bael/gatherData"
	"github.com/codegold79/bael/userData"
)

func main() {
	baseUrl := "https://www.ltd.org/system-map/"
	allAlerts, err := gatherData.ScrapeSite(baseUrl)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nwriting to file")
	gatherData.SaveAlertsToFile(allAlerts)
	fmt.Println("\nwriting to database")
	gatherData.SaveAlertsToDb(allAlerts)

	userKeys, err := userData.GetUserKeys()
	fmt.Println("User keys retrieved")
	if err != nil {
		fmt.Println(err)
	}

	for _, key := range userKeys {
		userData.RemoveOutdatedAlerts(key)
		fmt.Println("Outdated users' alerts removed")
		
		// Gather new alerts
		fmt.Println(userData.GatherNewUserAlerts(key))
		fmt.Println("Retrieved userData")
		
		// Send email with new alerts
		// Save keys in user data as they are no longer new
	}
}
