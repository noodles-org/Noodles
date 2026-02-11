package main

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fetch_ip_ops_processed_total",
		Help: "The total number of processed DNS records",
	})
	opsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fetch_ip_ops_failed_total",
	})
)

func getPublicIP() (string, error) {
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing url request:", err)
		}
	}()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func updateDNSRecords(client *cloudflare.Client, ip string, recordID string) error {
	_, err := client.DNS.Records.Edit(context.Background(), recordID,
		dns.RecordEditParams{
			ZoneID: cloudflare.F(os.Getenv("ZONE_ID")),
			Body: dns.ARecordParam{
				Content: cloudflare.F(ip),
			},
		})
	return err
}

func main() {
	ip, err := getPublicIP()
	if err != nil {
		opsFailed.Inc()
		panic(err)
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(os.Getenv("API_KEY")),
	)

	records := strings.Split(os.Getenv("RECORDS"), ",")
	for _, record := range records {
		recordID := strings.TrimSpace(record)
		err = updateDNSRecords(client, ip, recordID)
		if err != nil {
			opsFailed.Inc()
			panic(err)
		} else {
			opsProcessed.Inc()
		}
	}

	err = push.New(os.Getenv("SERVICE_URL"), "fetch-ip").
		Collector(opsProcessed).
		Collector(opsFailed).
		Push()
	if err != nil {
		fmt.Printf("Could not push metrics to Pushgateway: %v", err)
	}
}
