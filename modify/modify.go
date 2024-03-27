package modify

import (
	"cloud.google.com/go/logging"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"log"
	"strconv"
	"strings"
)

const secondsInYear = 31536000

type App struct {
	logger        *logging.Logger
	storageClient *storage.Client
}

func NewApp(logger *logging.Logger, bucket *storage.Client) *App {
	return &App{
		logger:        logger,
		storageClient: bucket,
	}
}

func GcsUploadEvent(a *App) func(context.Context, cloudevents.Event) error {
	return func(ctx context.Context, event cloudevents.Event) error {
		defer func() {
			err := a.logger.Flush()
			if err != nil {
				log.Printf("failed to flush logger %v", err)
			}
		}()
		eventData, err := extractEventData(event)
		if err != nil {
			a.logger.Log(logging.Entry{
				Severity: logging.Error,
				Payload:  fmt.Sprintf("failed to extract event data - %v", err),
			})
			return err
		}
		err = validateEventData(eventData)
		if err != nil {
			a.logger.Log(logging.Entry{
				Severity: logging.Error,
				Payload:  fmt.Sprintf("invalid event data - %v", err),
			})
			return err
		}

		a.logger.Log(logging.Entry{
			Severity: logging.Debug,
			Payload:  fmt.Sprintf("processing object %s/%s", eventData.Bucket, eventData.Name),
		})

		if !strings.Contains(eventData.Name, "images/") {
			a.logger.Log(logging.Entry{
				Severity: logging.Info,
				Payload:  fmt.Sprintf("skipping object %s/%s", eventData.Bucket, eventData.Name),
			})
			return nil
		}

		bucket := a.storageClient.Bucket(eventData.Bucket)
		object := bucket.Object(eventData.Name)
		object.BucketName()
		_, err = object.Attrs(ctx)
		if err != nil {
			if err.Error() == "storage: object doesn't exist" {
				a.logger.Log(logging.Entry{
					Severity: logging.Info,
					Payload:  fmt.Sprintf("object %s/%s doesn't exist", eventData.Bucket, eventData.Name),
				})

				// Object was probably deleted, we don't need to worry about it anymore.
				return nil
			}

			a.logger.Log(logging.Entry{
				Severity: logging.Error,
				Payload: fmt.Sprintf(
					"failed reading object %s/%s attributes - %v",
					eventData.Bucket,
					eventData.Name,
					err,
				),
			})
			return err
		}

		_, err = object.Update(ctx, storage.ObjectAttrsToUpdate{
			CacheControl: fmt.Sprintf("public, max-age=%s, immutable", strconv.Itoa(secondsInYear)),
		})
		if err != nil {
			a.logger.Log(logging.Entry{
				Severity: logging.Error,
				Payload: fmt.Sprintf(
					"failed updating object %s/%s - %v",
					eventData.Bucket,
					eventData.Name,
					err,
				),
			})
			return err
		}
		a.logger.Log(logging.Entry{
			Severity: logging.Info,
			Payload:  fmt.Sprintf("metadata set for object %s/%s", eventData.Bucket, eventData.Name),
		})

		return nil
	}
}
