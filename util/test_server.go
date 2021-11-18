package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// use this minimal http server to test the exporter locally
// Run the exporter with the flag --metadata-endpoint = "http://localhost:9092/latest/meta-data/"
func main() {
	http.HandleFunc("/latest/meta-data/spot/instance-action", func(w http.ResponseWriter, r *http.Request) {
		terminationTime := time.Now().Add(2 * time.Minute)
		utc, _ := time.LoadLocation("UTC")
		fmt.Fprintf(w, "{\"action\": \"stop\", \"time\": \"%s\"}", terminationTime.In(utc).Format(time.RFC3339))
	})

	http.HandleFunc("/latest/meta-data/instance-id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "i-0d2aab13057917887")
	})
	http.HandleFunc("/latest/meta-data/instance-type", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "c5.9xlarge")
	})
	http.HandleFunc("/latest/meta-data/events/recommendations/rebalance", func(w http.ResponseWriter, r *http.Request) {
		noticeTime := time.Now()
		utc, _ := time.LoadLocation("UTC")
		fmt.Fprintf(w, "{\"noticeTime\":\"%s\"}", noticeTime.In(utc).Format(time.RFC3339))
	})

	log.Fatal(http.ListenAndServe(":9092", nil))

}
