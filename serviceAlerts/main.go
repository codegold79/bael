package main

import (
	"fmt"
	"github.com/codegold79/bael/gatherData"
)

func main() {
	baseUrl := "https://www.ltd.org/system-map/"
	allAlerts, err := gatherData.ScrapeSite(baseUrl)

	if err != nil {
		fmt.Println(err)
	}

	gatherData.SaveAlertsToFile(allAlerts)
	gatherData.SaveAlertsToDb(allAlerts)
}
