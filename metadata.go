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
	metadataEndpoint          string
	rebalanceIndicator        *prometheus.Desc
	rebalanceScrapeSuccessful *prometheus.Desc
	scrapeSuccessful          *prometheus.Desc
	terminationIndicator      *prometheus.Desc
	terminationTime           *prometheus.Desc
}

type InstanceAction struct {
	Action string    `json:"action"`
	Time   time.Time `json:"time"`
}

type InstanceEvent struct {
	NoticeTime time.Time `json:"noticeTime"`
}

func NewTerminationCollector(me string) *terminationCollector {
	return &terminationCollector{
		metadataEndpoint:          me,
		rebalanceIndicator:        prometheus.NewDesc("aws_instance_rebalance_recommended", "Instance rebalance is recommended", []string{"instance_id", "instance_type"}, nil),
		rebalanceScrapeSuccessful: prometheus.NewDesc("aws_instance_metadata_service_events_available", "Metadata service events endpoint available", []string{"instance_id"}, nil),
		scrapeSuccessful:          prometheus.NewDesc("aws_instance_metadata_service_available", "Metadata service available", []string{"instance_id"}, nil),
		terminationIndicator:      prometheus.NewDesc("aws_instance_termination_imminent", "Instance is about to be terminated", []string{"instance_action", "instance_id", "instance_type"}, nil),
		terminationTime:           prometheus.NewDesc("aws_instance_termination_in", "Instance will be terminated in", []string{"instance_id", "instance_type"}, nil),
	}
}

func (c *terminationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rebalanceIndicator
	ch <- c.rebalanceScrapeSuccessful
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

	typeResp, err := client.Get(c.metadataEndpoint + "instance-type")
	var instanceType string
	if err != nil {
		log.Errorf("couldn't parse instance-type from metadata: %s", err.Error())
		return
	}
	if typeResp.StatusCode == 404 {
		log.Errorf("couldn't parse instance-type from metadata: endpoint not found")
		return
	}
	defer typeResp.Body.Close()
	body, _ = ioutil.ReadAll(typeResp.Body)
	instanceType = string(body)

	resp, err := client.Get(c.metadataEndpoint + "spot/instance-action")
	if err != nil {
		log.Errorf("Failed to fetch data from metadata service: %s", err)
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 0, instanceId)
	} else {
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 1, instanceId)

		if resp.StatusCode == 404 {
			log.Debug("instance-action endpoint not found")
			ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, "", instanceId, instanceType)
		} else {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)

			var ia = InstanceAction{}
			err := json.Unmarshal(body, &ia)

			// value may be present but not be a time according to AWS docs,
			// so parse error is not fatal
			if err != nil {
				log.Errorf("Couldn't parse instance-action metadata: %s", err)
				ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, instanceId, instanceType)
			} else {
				log.Infof("instance-action endpoint available, termination time: %v", ia.Time)
				ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 1, ia.Action, instanceId, instanceType)
				delta := ia.Time.Sub(time.Now())
				if delta.Seconds() > 0 {
					ch <- prometheus.MustNewConstMetric(c.terminationTime, prometheus.GaugeValue, delta.Seconds(), instanceId, instanceType)
				}
			}
		}
	}

	eventResp, err := client.Get(c.metadataEndpoint + "events/recommendations/rebalance")
	if err != nil {
		log.Errorf("Failed to fetch events data from metadata service: %s", err)
		ch <- prometheus.MustNewConstMetric(c.rebalanceScrapeSuccessful, prometheus.GaugeValue, 0, instanceId)
		// Return early as this is the last metric/metadata scrape attempt
		return
	} else {
		ch <- prometheus.MustNewConstMetric(c.rebalanceScrapeSuccessful, prometheus.GaugeValue, 1, instanceId)

		if eventResp.StatusCode == 404 {
			log.Debug("rebalance endpoint not found")
			ch <- prometheus.MustNewConstMetric(c.rebalanceIndicator, prometheus.GaugeValue, 0, instanceId, instanceType)
			// Return early as this is the last metric/metadata scrape attempt
			return
		} else {
			defer eventResp.Body.Close()
			body, _ := ioutil.ReadAll(eventResp.Body)

			var ie = InstanceEvent{}
			err := json.Unmarshal(body, &ie)

			if err != nil {
				log.Errorf("Couldn't parse rebalance recommendation event metadata: %s", err)
				ch <- prometheus.MustNewConstMetric(c.rebalanceIndicator, prometheus.GaugeValue, 0, instanceId, instanceType)
			} else {
				log.Infof("rebalance recommendation event endpoint available, recommendation time: %v", ie.NoticeTime)
				ch <- prometheus.MustNewConstMetric(c.rebalanceIndicator, prometheus.GaugeValue, 1, instanceId, instanceType)
			}
		}
	}
}
