package main

import (
	//"golang.org/x/net/html"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	url := "https://www.ltd.org/service-alerts/"
	wp, err := retrieveWebpage(url)

	if err != nil {
		wp = ""
		fmt.Print(err)
	}

	err = savePageToFile(wp)

	if err != nil {
		fmt.Print(err)
	}
}

func retrieveWebpage(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func savePageToFile(contents string) error {
	f, err := os.Create("outputs/ltd-service-alerts.txt")

	if err != nil {
		return err
	}

	defer f.Close()

	f.WriteString(contents)

	return nil
}
