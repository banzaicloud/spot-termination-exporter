package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type terminationCollector struct {
	metadataEndpoint     string
	scrapeSuccessful     *prometheus.Desc
	terminationIndicator *prometheus.Desc
	terminationTime      *prometheus.Desc
}

type InstanceAction struct {
	Action string    `json:"action"`
	Time   time.Time `json:"time"`
}

func NewTerminationCollector(me string) *terminationCollector {
	return &terminationCollector{
		metadataEndpoint:     me,
		scrapeSuccessful:     prometheus.NewDesc("aws_instance_metadata_service_available", "Metadata service available", []string{"instance_id"}, nil),
		terminationIndicator: prometheus.NewDesc("aws_instance_termination_imminent", "Instance is about to be terminated", []string{"instance_action", "instance_id"}, nil),
		terminationTime:      prometheus.NewDesc("aws_instance_termination_in", "Instance will be terminated in", []string{"instance_id"}, nil),
	}
}

func (c *terminationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scrapeSuccessful
	ch <- c.terminationIndicator
	ch <- c.terminationTime

}

func (c *terminationCollector) Collect(ch chan<- prometheus.Metric) {
	log.Info("Fetching termination data from metadata-service")
	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	idResp, err := client.Get(c.metadataEndpoint + "instance-id")
	var instanceId string
	if err != nil {
		log.Errorf("couldn't parse instance-id from metadata: %s", err.Error())
		return
	}
	if idResp.StatusCode == 404 {
		log.Errorf("couldn't parse instance-id from metadata: endpoint not found")
		return
	}
	defer idResp.Body.Close()
	body, _ := ioutil.ReadAll(idResp.Body)
	instanceId = string(body)

	resp, err := client.Get(c.metadataEndpoint + "spot/instance-action")
	if err != nil {
		log.Errorf("Failed to fetch data from metadata service: %s", err)
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 0, instanceId)
		return
	} else {
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 1, instanceId)

		if resp.StatusCode == 404 {
			log.Debug("instance-action endpoint not found")
			ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, "", instanceId)
			return
		} else {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)

			var ia = InstanceAction{}
			err := json.Unmarshal(body, &ia)

			// value may be present but not be a time according to AWS docs,
			// so parse error is not fatal
			if err != nil {
				log.Errorf("Couldn't parse instance-action metadata: %s", err)
				ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, instanceId)
			} else {
				log.Infof("instance-action endpoint available, termination time: %v", ia.Time)
				ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 1, ia.Action, instanceId)
				delta := ia.Time.Sub(time.Now())
				if delta.Seconds() > 0 {
					ch <- prometheus.MustNewConstMetric(c.terminationTime, prometheus.GaugeValue, delta.Seconds(), instanceId)
				}
			}
		}
	}
}
