package modify_gcs_uploads_func

import (
	"cloud.google.com/go/logging"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/itmayziii/modify_gcs_uploads_func/modify"
	"log"
	"os"
)

// init exists purely as an entry point for GCP Cloud functions.
// https://cloud.google.com/functions/docs/writing#entry-point
func init() {
	ctx := context.Background()
	loggingClient, err := logging.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatal("failed to create logging client", err)
	}
	logger := loggingClient.Logger("modify-gcs-upload-func", logging.RedirectAsJSON(os.Stdout))

	client, err := storage.NewClient(ctx)
	if err != nil {
		err := logger.LogSync(ctx, logging.Entry{
			Severity: logging.Alert,
			Payload:  fmt.Sprintf("failed to create storage client - %v", err),
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	functions.CloudEvent("ModifyGcsUpload", modify.GcsUploadEvent(modify.NewApp(
		logger,
		client,
	)))
}
