package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/alexruf/tankerkoenig-go"
	flags "github.com/jessevdk/go-flags"
	"github.com/uhlig-it/tankerkoenig-influxdb-importer/influxpublisher"
)

// ldflags will be set by goreleaser
var version = "vDEV"
var commit = "NONE"
var date = "UNKNOWN"

var options struct {
	Verbose []bool `short:"v" long:"verbose" description:"show verbose output. Repeat for even verboser output."`
	Version bool   `short:"V" long:"version" description:"show program version"`
}

func main() {
	p := flags.NewParser(&options, flags.Default)
	p.ShortDescription = "Exports Tankerkönig data to InfluxDB"
	p.LongDescription = "Fetches the current prices from Tankerkönig and writes them to an InfluxDB instance."

	ids, err := p.Parse()

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if options.Version {
		fmt.Printf("%s %s (%s), built on %s\n", getProgramName(), version, commit, date)
		os.Exit(0)
	}

	apiKey, present := os.LookupEnv("TANKERKOENIG_API_KEY")

	if !present {
		fmt.Fprintln(os.Stderr, "Error: Required TANKERKOENIG_API_KEY not present")
		os.Exit(1)
	}

	client := tankerkoenig.NewClient(apiKey, nil)

	if len(ids) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No station ID given.")
		os.Exit(1)
	}

	if len(ids) > 10 {
		fmt.Fprintln(os.Stderr, "Warning: More than 10 station IDs given; remainder will be ignored.")
	}

	influxURLValue, present := os.LookupEnv("INFLUXDB_URL")

	if !present {
		fmt.Fprintln(os.Stderr, "Error: Required INFLUXDB_URL not present")
		os.Exit(1)
	}

	influxURL, err := url.Parse(influxURLValue)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	publisher, err := influxpublisher.New(influxURL)

	if err != nil {
		fmt.Println("Error creating new InfluxDB publisher: ", err)
		os.Exit(1)
	}

	for _, id := range ids {
		ts := time.Now()
		station, _, err := client.Station.Detail(id)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not get details for station %s: %s", id, err)
			continue
		}

		fmt.Printf("%s %s (%v)\n", station.Name, station.Place, station.Id)
		fmt.Printf("  Time:   %v\n", ts.Format(time.RFC3339))
		fmt.Printf("  Open:   %v\n", station.IsOpen)
		fmt.Printf("  Diesel: %v €/l\n", station.Diesel)
		fmt.Printf("  E5:     %v €/l\n", station.E5)
		fmt.Printf("  E10:    %v €/l\n", station.E10)

		publisher.Publish(station, ts)
	}
}

func getProgramName() string {
	path, err := os.Executable()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not determine program name; using 'unknown'.")
		return "unknown"
	}

	return filepath.Base(path)
}
