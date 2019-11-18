# bael
## LTD-Service Alerts Project
* This is a project where a webpage (ltd.org) will be scraped for service alerts. Users will be able to sign up to receive SMS or email messages whenever the service alerts on the site changes, and they have subscribed to a route.

## Programming languages to be used
* backend will be in Go
* (not yet written) front end will be in Elm

# Some things this app does
* pull web page content with 
** "net/http", and
** "golang.org/x/net/html"
* Parse html with
** "github.com/PuerkitoBio/goquery"
* parse the data and save to a map with the help of "string"
* save the map contents using "os"
* search for excessive white space with regexp and replace them with the first match
* save data to Google Cloud Platform's Firestore database

# Future plans
* I'd like the program to be run by GCP Cloud Functions eventually.
