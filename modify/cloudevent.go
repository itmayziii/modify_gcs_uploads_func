package modify

import (
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// EventData is this applications specific event payload data needed to modify the GCS object.
type EventData struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// extractEventData unmarshals the event payload into our expected [EventData] format.
func extractEventData(event cloudevents.Event) (EventData, error) {
	var eventData EventData
	if err := event.DataAs(&eventData); err != nil {
		return EventData{}, err
	}
	return eventData, nil
}

// validateEventData ensures that [EventData] contains appropriate values such as having a valid bucket, name, etc...
func validateEventData(eventData EventData) error {
	if eventData.Bucket == "" {
		return errors.New("missing \"bucket\"")
	}

	if eventData.Name == "" {
		return errors.New("missing \"name\"")
	}

	return nil
}
