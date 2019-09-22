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

	gatherData.SaveAlertsToFile(allAlerts)
	gatherData.SaveAlertsToDb(allAlerts)

	users, err := userData.GetUserKeys()
	fmt.Println("User keys retrieved")
	if err != nil {
		fmt.Println(err)
	}

	for _, u := range users {
		userData.RemoveOutdatedAlerts(u)
		fmt.Println("Outdated users' alerts removed")
	}
}
