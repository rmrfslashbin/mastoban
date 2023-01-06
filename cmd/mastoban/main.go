package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rmrfslashbin/mastoban/pkg/geoip"
	"github.com/rmrfslashbin/mastoban/pkg/mastoclient"
	"github.com/rs/zerolog"
)

const (
	// APP_NAME is the name of the application
	APP_NAME = "mastoban"
)

// Context is used to pass context/global configs to the commands
type Context struct {
	// log is the logger
	log *zerolog.Logger
}

// LookupCmd runs an IP address lookup through the GeoIP database
type LookupCmd struct {
	IP     string  `required:"" name:"ip" help:"IP address to parse."`
	DBFile *string `name:"dbfile" env:"DBFILE" help:"Path to the GeoIP country database file."`
}

// Run is the entry point for LookupCmd command
func (r *LookupCmd) Run(ctx *Context) error {
	// Parse the IP address from the cli args
	ip := net.ParseIP(r.IP)
	if ip == nil {
		return errors.New("invalid IP address")
	}

	// Create a new GeoIP DB instance
	geoIP, err := geoip.New(r.DBFile)
	if err != nil {
		return err
	}

	// Lookup the IP addr in the GeoIP DB
	ipData, err := geoIP.Lookup(ip)
	if err != nil {
		return err
	}

	// Display a note if the IP is not in the US
	if ipData.Country != "US" {
		fmt.Println("*** IP address is not based in the US ***")
	}

	// Print the details from GeoIP DB lookup
	fmt.Printf("IP Addr:   %s\n", ipData.IP)
	fmt.Printf("Continent: %s\n", ipData.Continent)
	fmt.Printf("Country:   %s\n", ipData.Country)
	fmt.Println()

	return nil
}

// SuspendCmd is the command to suspend an account
type SuspendCmd struct {
	ID           string `required:"" name:"id" help:"ID of the account to suspend."`
	Instance     string `required:"" name:"instance" help:"Instance to suspend the account on."`
	AccessToken  string `required:"" name:"token" help:"Access token to use to suspend the account."`
	SuspendText  string `name:"text" default:"This accound is suspended pending further review." help:"Text to use when suspending the account."`
	SuspendLevel string `required:"" name:"level" env:"SUSPENDLEVEL" enum:"none,sensitive,disable,silence,suspend" help:"Suspend level to use when suspending the account."`
}

// Run is the entry point for SuspendCmd command
func (r *SuspendCmd) Run(ctx *Context) error {
	// Create a new Mastoclient instance
	mastodonClient, err := mastoclient.New(
		mastoclient.WithInstance(r.Instance),       // Instance URL from CLI args
		mastoclient.WithAccessToken(r.AccessToken), // Access Token from CLI args
		mastoclient.WithLogger(ctx.log),
	)
	if err != nil {
		return err
	}

	// Run the suspend funciton on the ID provides in CLI args
	err = mastodonClient.Suspend(
		&mastoclient.SuspendInput{
			ID:           r.ID,
			SuspendLevel: r.SuspendLevel,
			SuspendText:  r.SuspendText,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Suspended account %s on %s\n", r.ID, r.Instance)

	return nil
}

// CLI is the main CLI struct
type CLI struct {
	// Global flags/args
	LogLevel string `name:"loglevel" env:"LOGLEVEL" default:"info" enum:"panic,fatal,error,warn,info,debug,trace" help:"Set the log level."`

	Lookup  LookupCmd  `cmd:"" help:"Parse an IP address and look it up in the GeoIP database."`
	Suspend SuspendCmd `cmd:"" help:"Suspend an account."`
}

func main() {
	var err error

	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Parse the command line
	var cli CLI
	ctx := kong.Parse(&cli)

	// Set up the logger's log level
	// Default to info via the CLI args
	switch cli.LogLevel {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	// Log some start up stuff for debugging
	log.Debug().Msg("Starting up")

	// Call the Run() method of the selected parsed command.
	err = ctx.Run(&Context{log: &log})

	// FatalIfErrorf terminates with an error message if err != nil
	ctx.FatalIfErrorf(err)
}
