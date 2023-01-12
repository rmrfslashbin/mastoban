package app

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/davecgh/go-spew/spew"
	"github.com/rmrfslashbin/mastoban/pkg/queue"
	"github.com/rmrfslashbin/mastoban/pkg/structs"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func WebhookHandler(ctx context.Context, request events.APIGatewayProxyRequest) (*structs.Output, error) {
	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	queryPSK, ok := request.QueryStringParameters["psk"]
	if !ok {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "request.QueryStringParameters['psk']").
			Str("errRef", guid.String()).
			Str("QueryStringParameters", spew.Sdump(request.QueryStringParameters)).
			Msg("psk query param missing")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorMissingPSK(),
			},
		}, nil
	}

	// Fetch the PSK from the environment
	expectedPSK := os.Getenv("PSK")
	if expectedPSK == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('PSK')").
			Str("errRef", guid.String()).
			Msg("Failed to get PSK from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("PSK"),
			},
		}, nil
	}

	if queryPSK != expectedPSK {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "if queryPSK != expectedPSK").
			Str("errRef", guid.String()).
			Str("queryPSK", queryPSK).
			Str("expectedPSK", expectedPSK).
			Msg("Invalid PSK")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorPSKMismatch(),
			},
		}, nil
	}

	// Parse the request body
	message := &structs.AccoutCreatedEvent{}
	err := json.Unmarshal([]byte(request.Body), message)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "json.Unmarshal([]byte(request.Body), message)").
			Str("errRef", guid.String()).
			Str("Message", string(request.Body)).
			Msg("Failed to unmarshal request body")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToUnmarshalRequest(),
			},
		}, nil
	}

	if message.Event != "account.created" {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "message.Event check").
			Str("errRef", guid.String()).
			Str("MessageEvent", message.Event).
			Msg("Message event is not supported")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorMessageEventNotSupported(),
			},
		}, nil
	}

	// Fetch the PSK from the environment
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	if expectedPSK == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('SQS_QUEUE_URL')").
			Str("errRef", guid.String()).
			Msg("Failed to get SQS queue URL from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("SQS_QUEUE_URL"),
			},
		}, nil
	}

	sqs, err := queue.New(
		queue.WithLogger(&log),
		queue.WithSQSURL(sqsQueueURL),
	)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "queue.New()").
			Str("errRef", guid.String()).
			Msg("Failed to create SQS queue instance")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToCreateQueueInstance(),
			},
		}, nil
	}

	if err := sqs.SendWorkerMessage(message); err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "sqs.SendWorkerMessage(message)").
			Str("errRef", guid.String()).
			Msg("Failed to send message to SQS queue")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToSendMessageToQueue(),
			},
		}, nil
	}

	return &structs.Output{
		Status: "accepted",
	}, nil
}
