package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// use this minimal http server to test the exporter locally
// change constant to have metadataEndpoint = "http://localhost:9092/latest/meta-data/spot/instance-action"
func main() {
	http.HandleFunc("/latest/meta-data/spot/instance-action", func(w http.ResponseWriter, r *http.Request) {
		terminationTime := time.Now().Add(2 * time.Minute)
		utc, _ := time.LoadLocation("UTC")
		fmt.Fprintf(w, "{\"action\": \"stop\", \"time\": \"%s\"}", terminationTime.In(utc).Format(time.RFC3339))
	})

	http.HandleFunc("/latest/meta-data/instance-id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "i-0d2aab13057917887")
	})

	log.Fatal(http.ListenAndServe(":9092", nil))

}
