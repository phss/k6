package cloud

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/loadimpact/k6/stats"
)

type Collector struct {
	seenMetrics []string
	referenceID string

	client *Client
}

func New(fname string) (*Collector, error) {
	referenceID := os.Getenv("K6CLOUD_REFERENCEID")

	return &Collector{
		seenMetrics: []string{},
		referenceID: referenceID,
		client:      NewClient("token"),
	}, nil
}

func (c *Collector) Run(ctx context.Context) {
	t := time.Now()
	<-ctx.Done()
	s := time.Now()

	log.Debug(fmt.Sprintf("http://localhost:5000/v1/metrics/%s/%d000/%d000\n", c.referenceID, t.Unix(), s.Unix()))
}

func (c *Collector) Collect(samples []stats.Sample) {
	var cloudSamples []*Sample
	for _, sample := range samples {
		if c.HasSeenMetric(sample.Metric) {
			continue
		}

		sampleJSON := &Sample{
			Type:   "Point",
			Metric: sample.Metric.Name,
			Data: SampleData{
				Time:  sample.Time,
				Value: sample.Value,
				Tags:  sample.Tags,
			},
		}
		cloudSamples = append(cloudSamples, sampleJSON)

		c.seenMetrics = append(c.seenMetrics, sample.Metric.Name)
	}

	if len(cloudSamples) > 0 {
		c.client.PushMetric(c.referenceID, cloudSamples)
	}
}

func (c *Collector) HasSeenMetric(m *stats.Metric) bool {
	for _, n := range c.seenMetrics {
		if n == m.Name {
			return true
		}
	}
	return false
}
