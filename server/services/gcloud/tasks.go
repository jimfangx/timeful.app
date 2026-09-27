package gcloud

import (
	"context"
	"os"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2beta3"
	"cloud.google.com/go/cloudtasks/apiv2beta3/cloudtaskspb"
	"google.golang.org/api/option"
	"schej.it/server/logger"
)

var TasksClient *cloudtasks.Client

func InitTasks() func() {
	credsFile := os.Getenv("SERVICE_ACCOUNT_KEY_PATH")
	if credsFile == "" || credsFile == "?" {
		logger.StdOut.Println("SERVICE_ACCOUNT_KEY_PATH not set, Cloud Tasks disabled")
		return func() {}
	}

	ctx := context.Background()
	var err error
	TasksClient, err = cloudtasks.NewClient(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		logger.StdErr.Println("Failed to initialize Cloud Tasks:", err)
		return func() {}
	}

	// Return function to close client
	return func() {
		TasksClient.Close()
	}
}

// CreateEmailTask previously scheduled the 0h/24h/3-day "you haven't
// responded yet" reminder emails via Listmonk. That reminder flow has been
// disabled (as part of the Mailgun migration, since it relied on Cloud Tasks
// calling Listmonk's HTTP endpoint directly), so this is now a no-op that
// returns no task IDs. DeleteEmailTask is kept working as-is so it can still
// cancel any reminder tasks that were already scheduled before this change.
func CreateEmailTask(email string, ownerName string, eventName string, eventId string) []string {
	return []string{}
}

func DeleteEmailTask(taskId string) {
	if TasksClient == nil {
		logger.StdErr.Println("WARNING: Cloud Tasks is disabled, skipping DeleteEmailTask")
		return
	}

	err := TasksClient.DeleteTask(context.Background(), &cloudtaskspb.DeleteTaskRequest{
		Name: taskId,
	})

	if err != nil {
		// logger.StdErr.Println(err)
		return
	}
}
