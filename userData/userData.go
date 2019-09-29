package userData

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
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

type UserInfo struct {
	Route_ids         []string `firestore:"route_ids"`
	Stored_alert_keys []string `firestore:"stored_alert_keys"`
	Email             string   `firestore:"email"`
	Name              string   `firestore:"name"`
}

func GatherUserInfo(userKey string) (UserInfo, error) {
	var userInfo UserInfo
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		return userInfo, err
	}

	defer client.Close()

	userInfo, err = GetUserInfo(userKey)
	fmt.Printf("Retrieved alerts for userKey: %v\n", userKey)

	if err != nil {
		return userInfo, err
	}

	return userInfo, nil
}

func GatherUserNewAlerts(userKey string, userRoutes []string, userAlerts []string) ([]string, error) {
	var routeAlerts []string
	var newAlerts []string

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")
	if err != nil {
		return newAlerts, err
	}

	for _, ur := range userRoutes {
		iter := client.Collection("alerts").Where("outdated_at", "==", nil).Where("route_ids", "array-contains", ur).Documents(ctx)
		for {
			doc, err := iter.Next()

			if err == iterator.Done {
				break
			}

			routeAlerts = append(routeAlerts, doc.Ref.ID)
		}
	}

	var match bool
	for _, ra := range routeAlerts {
		match = false
		for _, ua := range userAlerts {
			if ra == ua {
				match = true
			}
		}
		if !match {
			newAlerts = append(newAlerts, ra)
		}
	}

	return newAlerts, nil
}

// Send update emails
// if the email array is not empty, send emails to the user's email address
func SendAlertsToUserEmail(userKey string, alerts []string) error {
	return nil
}

// Return the routes that a user is subscribed to.
func GetUserInfo(userKey string) (UserInfo, error) {
	var userInfo UserInfo

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "ltd-sched-mon")

	if err != nil {
		return userInfo, err
	}

	defer client.Close()

	doc, err := client.Collection("users").Doc(userKey).Get(ctx)

	if err != nil {
		return userInfo, err
	}

	err = doc.DataTo(&userInfo)

	if err != nil {
		return userInfo, err
	}

	return userInfo, nil
}
