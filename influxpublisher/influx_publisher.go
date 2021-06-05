package influxpublisher

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/alexruf/tankerkoenig-go"
	influx "github.com/influxdata/influxdb/client/v2"
)

type InfluxPublisher struct {
	influxClient influx.Client
	database     string
}

func New(influxURL *url.URL) (*InfluxPublisher, error) {
	influxPassword, _ := influxURL.User.Password()
	influxDatabase := path.Base(influxURL.Path)

	if influxDatabase == "." || influxDatabase == "/" {
		fmt.Fprintln(os.Stderr, "Error: Database missing in InfluxDB URL.")
		os.Exit(1)
	}

	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     fmt.Sprintf("%s://%s", influxURL.Scheme, influxURL.Host),
		Username: influxURL.User.Username(),
		Password: influxPassword,
	})

	if err != nil {
		return nil, err
	}

	return &InfluxPublisher{influxClient: influxClient, database: influxDatabase}, nil
}

func (p *InfluxPublisher) Publish(station tankerkoenig.Station, ts time.Time) error {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  p.database,
		Precision: "s",
	})

	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"E5":     station.E5,
		"E10":    station.E10,
		"Diesel": station.Diesel,
	}

	tags := map[string]string{
		"station.id":       station.Id,
		"station.brand":    station.Brand,
		"station.is-open":  fmt.Sprintf("%v", station.IsOpen),
		"station.name":     station.Name,
		"station.place":    station.Place,
		"station.postcode": fmt.Sprintf("%v", station.PostCode),
	}

	pt, err := influx.NewPoint("price", tags, fields, ts)

	if err != nil {
		return err
	}

	bp.AddPoint(pt)

	return p.influxClient.Write(bp)
}
