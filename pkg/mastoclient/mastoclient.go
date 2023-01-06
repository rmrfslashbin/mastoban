package mastoclient

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Options for the mastoclient instance
type Option func(c *Config)

// Config for the mastoclient instance
type Config struct {
	log         *zerolog.Logger
	instance    string
	accessToken string
}

// New creates a new mastoclinet instance
func New(opts ...Option) (*Config, error) {
	c := &Config{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(c)
	}

	// Check to ensure instance is set
	if c.instance == "" {
		return nil, &NoInstance{}
	}

	// Check to ensure the access token is set
	if c.accessToken == "" {
		return nil, &NoAccessToken{}
	}

	// set up logger if not provided
	if c.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		c.log = &log
	}
	return c, nil
}

// WithInstance sets the instance to use
func WithInstance(instance string) Option {
	return func(c *Config) {
		c.instance = instance
	}
}

// WithAccessToken sets the access token to use
func WithAccessToken(accessToken string) Option {
	return func(c *Config) {
		c.accessToken = accessToken
	}
}

// WithLogger sets the logger to use
func WithLogger(log *zerolog.Logger) Option {
	return func(c *Config) {
		c.log = log
	}
}

type SuspendInput struct {
	ID           string
	SuspendText  string
	SuspendLevel string
}

// Suspend attempts to suspend a given Mastodon account
// and provides the user details via suspendText.
func (c *Config) Suspend(in *SuspendInput) error {
	suspendLevel := strings.ToLower(in.SuspendLevel)

	// Valid suspend types as defined by https://docs.joinmastodon.org/methods/admin/accounts/#form-data-parameters
	validTypes := make(map[string]struct{})
	validTypes["none"] = struct{}{}
	validTypes["sensitive"] = struct{}{}
	validTypes["disable"] = struct{}{}
	validTypes["silence"] = struct{}{}
	validTypes["suspend"] = struct{}{}

	// Check to ensure the suspend level is valid
	if _, ok := validTypes[suspendLevel]; !ok {
		return &InvalidSuspendType{typeProvided: &suspendLevel}
	}

	// Construct the API endpoint
	endpoint := c.instance + "/api/v1/admin/accounts/" + in.ID + "/action"

	// Set up required form key/value pairs
	data := url.Values{}
	data.Set("type", in.SuspendLevel)
	data.Set("text", in.SuspendText)
	data.Set("send_email_notification", "true")

	//create new POST request to the url and encoded form Data
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return err
	}

	// Set the required headers
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	//send request and get the response
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// body will contain failure condition text
	body, _ := io.ReadAll(res.Body)

	// On success, just return
	if res.StatusCode == 200 {
		return nil
	} else {
		// Otherwise, return an error message with the body/failure text
		return &PostFailed{
			Status: res.Status,
			Msg:    "Failed to suspend account " + in.ID,
			Err:    errors.New(string(body)),
		}
	}
}
