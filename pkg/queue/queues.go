package queue

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rmrfslashbin/mastoban/pkg/structs"
	"github.com/rs/zerolog"
)

type Option func(config *Config)

// Configuration structure.
type Config struct {
	sqsQueueURL string
	region      string
	profile     string
	log         *zerolog.Logger
	sqs         *sqs.Client
}

func New(opts ...func(*Config)) (*Config, error) {
	cfg := &Config{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.region == "" {
		cfg.region = os.Getenv("AWS_REGION")
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = cfg.region
		if cfg.profile != "" {
			o.SharedConfigProfile = cfg.profile
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// set up logger if not provided
	if cfg.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		cfg.log = &log
	}

	cfg.sqs = sqs.NewFromConfig(awsConfig)
	return cfg, nil
}

func WithLogger(log *zerolog.Logger) Option {
	return func(config *Config) {
		config.log = log
	}
}

func WithProfile(profile string) Option {
	return func(config *Config) {
		config.profile = profile
	}
}

func WithRegion(region string) Option {
	return func(config *Config) {
		config.region = region
	}
}

func WithSQSURL(sqsQueueURL string) Option {
	return func(config *Config) {
		config.sqsQueueURL = sqsQueueURL
	}
}

func (config *Config) SendWorkerMessage(event *structs.AccoutCreatedEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		config.log.Error().
			Str("process", "queues::SendWorkerMessage::json.Marshal()").
			Str("mastodon.userId", event.Object.Id).
			Str("mastodon.username", event.Object.Username).
			Str("mastodon.domian", event.Object.Domain).
			Str("mastodon.createdAt", event.Object.CreatedAt).
			Err(err).
			Msg("error marshalling event to JSON")
		return err
	}
	message := &sqs.SendMessageInput{
		QueueUrl:    aws.String(config.sqsQueueURL),
		MessageBody: aws.String(string(eventJSON)),
	}

	opt, err := config.sqs.SendMessage(context.TODO(), message)
	if err != nil {
		config.log.Error().
			Str("process", "queues::SendWorkerMessage::sqs.SendMessage()").
			Str("mastodon.userId", event.Object.Id).
			Str("mastodon.username", event.Object.Username).
			Str("mastodon.domian", event.Object.Domain).
			Str("mastodon.createdAt", event.Object.CreatedAt).
			Err(err).
			Msg("error sending message to SQS")
		return err
	}

	config.log.Info().
		Str("sqs.messageId", *opt.MessageId).
		Str("mastodon.userId", event.Object.Id).
		Str("mastodon.username", event.Object.Username).
		Str("mastodon.domian", event.Object.Domain).
		Str("mastodon.createdAt", event.Object.CreatedAt).
		Msg("sent message to SQS")
	return err
}

func (config *Config) GetAttribs() (map[string]string, error) {
	message := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(config.sqsQueueURL),
		AttributeNames: []types.QueueAttributeName{
			"All",
		},
	}
	if ret, err := config.sqs.GetQueueAttributes(context.TODO(), message); err != nil {
		return nil, err
	} else {
		return ret.Attributes, nil
	}
}

func (config *Config) Purge() error {
	message := &sqs.PurgeQueueInput{
		QueueUrl: aws.String(config.sqsQueueURL),
	}
	if _, err := config.sqs.PurgeQueue(context.TODO(), message); err != nil {
		return err
	}
	return nil
}
