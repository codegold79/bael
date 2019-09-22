// Send update emails
// loop through the all the user's routes
// 	in service_alerts, find all alerts that contain the route_id that have outdated_at = null
// loop through the found alerts
// 	if the route isn't already in the user's stored_service_alerts array, then store it in the email array
//  if the route is already in the user's stored service_alerts array, do nothing
// if the email array is not empty, send emails to the user's email address

package userData

import (
	"cloud.google.com/go/firestore"
	"context"
	"google.golang.org/api/iterator"
)

func GetUserKeys() ([]string, error) {
	var users []string

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		return users, err
	}

	defer client.Close()

	iter := client.Collection("users").Documents(ctx)

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		users = append(users, doc.Ref.ID)
	}

	return users, err
}

func RemoveOutdatedAlerts(userKey string) error {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		return err
	}

	defer client.Close()

	type UserInfo struct {
		Stored_alert_keys []string `firestore:"stored_alert_keys"`
	}
	var userInfo UserInfo

	// Retrieve user's alert keys.
	uDoc, err := client.Collection("users").Doc(userKey).Get(ctx)

	if err != nil {
		return err
	}

	// Get user's alert keys.
	err = uDoc.DataTo(&userInfo)

	if err != nil {
		return err
	}

	var currentAlerts []string

	// Find the alert keys from service_alerts collection.
	for _, key := range userInfo.Stored_alert_keys {
		aDoc, err := client.Collection("alerts").Doc(key).Get(ctx)

		if err != nil {
			return err
		}

		outdated_at, err := aDoc.DataAt("outdated_at")

		if outdated_at == nil {
			// If the alert is not outdated, add to a new slice.
			currentAlerts = append(currentAlerts, key)
		}
	}

	_, err = client.Collection("users").Doc(userKey).Set(ctx, map[string]interface{}{
		"stored_alert_keys": currentAlerts,
	}, firestore.MergeAll)

	return nil
}
