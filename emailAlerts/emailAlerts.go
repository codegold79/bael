package emailAlerts

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"os"
)

func SendEmail(recipientEmail string, alertKeys []string, currentDbAlerts map[string]string) error {
	from := mail.NewEmail("codegold test sender", "test@example.com")
	subject := "New LTD Service Alerts for Your Routes"
	to := mail.NewEmail("codegold recipient", recipientEmail)
	plainTextContent := AlertPlainTexts(alertKeys, currentDbAlerts)
	htmlContent := "<h3>Service Alerts</h3>" + AlertHtmlTexts(alertKeys, currentDbAlerts)
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

func AlertHtmlTexts(alertKeys []string, currentDbAlerts map[string]string) string {
	var str string
	for _, v := range alertKeys {
		str += "<h3>Key: " + v + "</h3>"
		str += "<p>" + currentDbAlerts[v] + "</p>"
	}
	return str
}

func AlertPlainTexts(alertKeys []string, currentDbAlerts map[string]string) string {
	var str string
	for _, v := range alertKeys {
		str += "================ Key: " + v + "===================== \n"
		str += currentDbAlerts[v] + "\n"
	}
	return str
}
