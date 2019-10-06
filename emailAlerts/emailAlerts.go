package emailAlerts

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"os"
)

func SendEmail(recipientEmail string, alertKeys []string, currentDbAlerts map[string][2]string) error {
	from := mail.NewEmail("Codegold79's LTD Service Alerts", "test@example.com")
	subject := "New LTD Service Alerts for Your Routes"
	to := mail.NewEmail("Codegold79 Email Recipient", recipientEmail)
	plainTextContent := AlertPlainTexts(alertKeys, currentDbAlerts)
	htmlContent := "<h2>Service Alerts</h2>" + AlertHtmlTexts(alertKeys, currentDbAlerts)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	response, err := client.Send(message)

	if err != nil {
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

	return nil
}

func AlertHtmlTexts(alertKeys []string, currentDbAlerts map[string][2]string) string {
	var str string
	for _, v := range alertKeys {
		str += "<h4>" + currentDbAlerts[v][1] + "</h4>"
		str += "<p>" + currentDbAlerts[v][0] + "</p>"
	}
	return str
}

func AlertPlainTexts(alertKeys []string, currentDbAlerts map[string][2]string) string {
	var str string
	for _, v := range alertKeys {
		str += "================ " + currentDbAlerts[v][0] + "===================== \n"
		str += currentDbAlerts[v][1] + "\n"
	}
	return str
}
