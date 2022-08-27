package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/alexruf/tankerkoenig-go"
	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ldflags will be set by goreleaser
var version = "vDEV"
var commit = "NONE"
var date = "UNKNOWN"

var options struct {
	Verbose            bool   `long:"verbose" short:"v" description:"show verbose output"`
	Version            bool   `long:"version" short:"V" description:"show program version"`
	MetricsBindAddress string `long:"bind-address" default:"localhost:9104" description:"bind address of the Prometheus metrics server"`
	Interval           string `long:"interval" short:"i" default:"15m" description:"fetch interval for price info"`
}

func main() {
	log.SetFlags(0) // no timestamp etc. - we have systemd's timestamps in the log anyway

	p := flags.NewParser(&options, flags.Default)
	p.ShortDescription = "Exports Tankerkönig data for Prometheus"
	p.LongDescription = "Fetches the current prices from Tankerkönig and exports them for Prometheus."

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
		log.Fatal("Error: Required TANKERKOENIG_API_KEY not present")
	}

	client := tankerkoenig.NewClient(apiKey, nil)

	if len(ids) == 0 {
		log.Fatal("Error: No station ID given.")
	}

	if len(ids) > 10 {
		log.Print("Warning: More than 10 station IDs given; remainder will be ignored.")
	}

	isOpenGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tankerkoenig_open_state",
		Help: "whether the station is open",
	}, []string{"station"})

	dieselGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tankerkoenig_diesel_euro_liter",
		Help: "price of Diesel, in €/l",
	}, []string{"station"})

	e5Gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tankerkoenig_e5_euro_liter",
		Help: "price of E5, in €/l",
	}, []string{"station"})

	e10Gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tankerkoenig_e10_euro_liter",
		Help: "price of E10, in €/l",
	}, []string{"station"})

	interval, err := time.ParseDuration(options.Interval)

	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	go func() {
		if options.Verbose {
			log.Printf("Checking prices every %v", interval)
		}

		for {
			select {
			case <-ticker.C:
				for _, id := range ids {
					if options.Verbose {
						log.Printf("Checking prices for id %v", id)
					}

					station, _, err := client.Station.Detail(id)

					if err != nil {
						log.Printf("Error: could not get details for station %s: %s", id, err)
						continue
					}

					stationLabel := fmt.Sprintf("%v %v", station.Name, station.Place)
					log.Printf("%s (%v)\n", stationLabel, station.Id)
					log.Printf("  Time:   %v\n", time.Now().Format(time.RFC3339))
					log.Printf("  Open:   %v\n", station.IsOpen)
					log.Printf("  Diesel: %v €/l\n", station.Diesel)
					log.Printf("  E5:     %v €/l\n", station.E5)
					log.Printf("  E10:    %v €/l\n", station.E10)

					publishAsBool(isOpenGauge, station.IsOpen, "Open", stationLabel)
					publishAsFloat(dieselGauge, station.Diesel, "Diesel", stationLabel)
					publishAsFloat(e5Gauge, station.E5, "E5", stationLabel)
					publishAsFloat(e10Gauge, station.E10, "E10", stationLabel)
				}
			case <-quit:
				if options.Verbose {
					log.Printf("Shutting down ticker")
				}

				ticker.Stop()

				return
			}
		}
	}()

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		if options.Verbose {
			log.Printf("Shutting down program")
		}

		close(quit)
		os.Exit(1)
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting metrics server at %v", options.MetricsBindAddress)
	log.Fatal(http.ListenAndServe(options.MetricsBindAddress, nil))
}

func getProgramName() string {
	path, err := os.Executable()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not determine program name; using 'unknown'.")
		return "unknown"
	}

	return filepath.Base(path)
}

func publishAsFloat(gauge *prometheus.GaugeVec, price interface{}, product string, stationLabel string) {
	switch convertedPrice := price.(type) {
	case float64:
		if options.Verbose {
			log.Printf("Publishing %v price %v at %v", product, stationLabel, convertedPrice)
		}

		gauge.WithLabelValues(stationLabel).Set(float64(convertedPrice))
	default:
		log.Printf("%v has no price for %v", stationLabel, product)
	}
}

func publishAsBool(gauge *prometheus.GaugeVec, state interface{}, product string, stationLabel string) {
	switch convertedState := state.(type) {
	case bool:
		if options.Verbose {
			log.Printf("Publishing %v state %v as %v", product, stationLabel, convertedState)
		}

		if convertedState {
			gauge.WithLabelValues(stationLabel).Set(1)
		} else {
			gauge.WithLabelValues(stationLabel).Set(0)
		}
	default:
		log.Printf("%v has no state for %v", stationLabel, product)
	}
}
