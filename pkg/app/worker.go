package app

import (
	"context"
	"encoding/json"
	"net"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rmrfslashbin/mastoban/pkg/geoip"
	"github.com/rmrfslashbin/mastoban/pkg/mastoclient"
	"github.com/rmrfslashbin/mastoban/pkg/structs"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// AppHandler is the entry point for the Lambda function
func WorkerHandler(ctx context.Context, request events.SQSEvent) (*structs.Output, error) {

	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// set up the config struct
	cfg := &config{}
	cfg.log = &log

	// Fetch the GeoIP database path from the environment
	geoIpDBPath := os.Getenv("GEOIP_DATABSE_PATH")
	if geoIpDBPath == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('GEOIP_DATABSE_PATH')").
			Str("errRef", guid.String()).
			Msg("Failed to get GEOIP_DATABSE_PATH from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("GEOIP_DATABSE_PATH"),
			},
		}, nil
	}

	// Sert up the GeoIP database instance
	geoIpDB, err := geoip.New(&geoIpDBPath)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "geoip.New()").
			Str("errRef", guid.String()).
			Str("GeoIPDBPath", geoIpDBPath).
			Msg("Failed to create new geoip instance")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToCreateGeoIPInstance(),
			},
		}, nil
	}

	// Fetch the Mastodon access token from the environment
	accessToken := os.Getenv("MASTODON_ACCESS_TOKEN")
	if accessToken == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('MASTODON_ACCESS_TOKEN')").
			Str("errRef", guid.String()).
			Msg("Failed to get MASTODON_ACCESS_TOKEN from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("MASTODON_ACCESS_TOKEN"),
			},
		}, nil
	}

	// Fetch the Mastodon instance URL from the environment
	instanceURL := os.Getenv("MASTODON_INSTANCE_URL")
	if instanceURL == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('MASTODON_INSTANCE_URL')").
			Str("errRef", guid.String()).
			Msg("Failed to get MASTODON_INSTANCE_URL from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("MASTODON_INSTANCE_URL"),
			},
		}, nil
	}

	// Fetch the Mastodon suspend text from the environment
	suspendText := os.Getenv("MASTODON_SUSPEND_TEXT")
	if suspendText == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('MASTODON_SUSPEND_TEXT')").
			Str("errRef", guid.String()).
			Msg("Failed to get MASTODON_SUSPEND_TEXT from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("MASTODON_SUSPEND_TEXT"),
			},
		}, nil
	}

	// none, sensitive, disable, silence, suspend
	suspendLevel := os.Getenv("MASTODON_SUSPEND_LEVEL")
	if suspendLevel == "" {
		guid := xid.New()
		log.Error().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "os.Getenv('MASTODON_SUSPEND_LEVEL')").
			Str("errRef", guid.String()).
			Msg("Failed to get MASTODON_SUSPEND_LEVEL from environment")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToFetchEnvVar("MASTODON_SUSPEND_LEVEL"),
			},
		}, nil
	}

	// Create a new mastoclient instance
	mastodonClient, err := mastoclient.New(
		mastoclient.WithInstance(instanceURL),
		mastoclient.WithAccessToken(accessToken),
		mastoclient.WithLogger(&log),
	)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "mastoclient.New()").
			Str("errRef", guid.String()).
			Msg("Failed to create new mastoclient instance")
		return &structs.Output{
			Error: &structs.Err{
				ErrRef: guid.String(), Msg: errorUnableToCreateMastoclientInstance(),
			},
		}, nil
	}

	impactedUsers := []structs.EventObject{}

	for i := range request.Records {
		message := &structs.AccoutCreatedEvent{}
		if err := json.Unmarshal([]byte(request.Records[i].Body), message); err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("module", MODULE).
				Str("function", "AppHandler").
				Str("process", "json.Unmarshal() request body").
				Str("errRef", guid.String()).
				Str("requestBody", string(request.Records[i].Body)).
				Msg("Failed to unmarshal request body")
			continue
		}

		// Parse the IP address from the request
		userIP := net.ParseIP(message.Object.Ip)
		if userIP == nil {
			guid := xid.New()
			log.Error().
				Str("module", MODULE).
				Str("function", "AppHandler").
				Str("process", "net.ParseIP()").
				Str("IP", message.Object.Ip).
				Str("errRef", guid.String()).
				Msg("Filed to parse IP address")
			continue
		}

		// Lookup the IP address in the GeoIP database
		ipData, err := geoIpDB.Lookup(userIP)
		if err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("module", MODULE).
				Str("function", "AppHandler").
				Str("process", "geoIpDB.Lookup()").
				Str("IP", userIP.String()).
				Str("errRef", guid.String()).
				Msg("Filed to lookup IP address in GeoIP database")
			continue
		}

		// If the IP is from the US, do nothing
		if ipData.Country == "US" {
			guid := xid.New()
			log.Info().
				Str("module", MODULE).
				Str("function", "AppHandler").
				Str("process", "geoIpDB.Lookup()").
				Str("errRef", guid.String()).
				Str("IP", ipData.IP.String()).
				Str("Country", ipData.Country).
				Str("Continent", ipData.Continent).
				Str("UserID", message.Object.Id).
				Str("Username", message.Object.Username).
				Str("Domain", message.Object.Domain).
				Str("Email", message.Object.Email).
				Str("CreatedAt", message.Object.CreatedAt).
				Msg("IP is from the US. Doing nothing.")
			continue
		}

		// Suspend the user!
		err = mastodonClient.Suspend(
			&mastoclient.SuspendInput{
				ID:           message.Object.Id,
				SuspendText:  suspendText,
				SuspendLevel: suspendLevel})
		if err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("module", MODULE).
				Str("function", "AppHandler").
				Str("process", "mastodonClient.Suspend()").
				Str("UserID", message.Object.Id).
				Str("errRef", guid.String()).
				Msg("Failed to suspend user")
			continue
		}

		// Log the details and return
		guid := xid.New()
		log.Info().
			Str("module", MODULE).
			Str("function", "AppHandler").
			Str("process", "geoIpDB.Lookup()").
			Str("errRef", guid.String()).
			Str("IP", ipData.IP.String()).
			Str("Country", ipData.Country).
			Str("Continent", ipData.Continent).
			Str("UserID", message.Object.Id).
			Str("Username", message.Object.Username).
			Str("Domain", message.Object.Domain).
			Str("Email", message.Object.Email).
			Str("CreatedAt", message.Object.CreatedAt).
			Msg("IP is not from the US. Account Suspended!")

		impactedUsers = append(impactedUsers, structs.EventObject{
			Username:  message.Object.Username,
			Id:        message.Object.Id,
			Domain:    message.Object.Domain,
			Email:     message.Object.Email,
			CreatedAt: message.Object.CreatedAt,
		})
	}
	return &structs.Output{
		Status: "suspended",
		Users:  &impactedUsers,
	}, nil
}
