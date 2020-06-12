package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type SpotinstResponce struct {
	Request struct {
		ID        string    `json:"id"`
		URL       string    `json:"url"`
		Method    string    `json:"method"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"request"`
	Response struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
		Kind  string `json:"kind"`
		Items []struct {
			InstanceID     string `json:"instanceId"`
			LifeCycleState string `json:"lifeCycleState"`
			PrivateIP      string `json:"privateIp"`
			GroupID        string `json:"groupId"`
		} `json:"items"`
		Count int `json:"count"`
	} `json:"response"`
}

func getSpotinstStatus(instance_id string) string {

	token := os.Getenv("SPOTINST_TOKEN")
	account_id := os.Getenv("SPOTINST_ACCOUNT_ID")

	if (token == "") || (account_id == "") {
		return "not_available"
	}

	url := fmt.Sprintf("https://api.spotinst.io/aws/ec2/instance/%s?accountId=%s", instance_id, account_id)

	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	var spotResp SpotinstResponce

	err = json.NewDecoder(resp.Body).Decode(&spotResp)
	if err != nil {
		log.Error(err)
	}

	if resp.StatusCode != 200 {
		log.Errorf("%v, %v", resp.StatusCode, resp.Status)
	}

	instanceStatus := spotResp.Response.Items[0].LifeCycleState
	if instanceStatus != "" {
		return instanceStatus
	}
	return "not_available"
}
