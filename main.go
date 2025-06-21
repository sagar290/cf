package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type DnsRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type DnsRecordsResponse struct {
	Result []DnsRecord `json:"result"`
}

var apiToken string
var newIP string
var proxied bool
var ttl int16 = 3600

func main() {
	rootCmd := &cobra.Command{Use: "cf"}

	updateCmd := &cobra.Command{
		Use:   "update dns [domain] [fqdn]",
		Short: "Update specific A record (like root or www) for a domain",
		RunE: func(cmd *cobra.Command, args []string) error {

			resource := args[0]
			domain := args[1]
			fqdn := args[2]

			if resource != "dns" {
				return fmt.Errorf("Invalid resource type: %s. Only 'dns' is supported", resource)
			}

			if fqdn == "@" {
				fqdn = domain
			}

			if apiToken == "" {
				apiToken = os.Getenv("CF_API_TOKEN")
			}

			if apiToken == "" {
				return fmt.Errorf("Cloudflare API token not provided")
			}

			if newIP == "" {
				return fmt.Errorf("New IP address must be provided via --ip flag")
			}

			return UpdateARecord(apiToken, domain, fqdn, newIP)
		},
	}

	updateCmd.Flags().StringVar(&newIP, "ip", "", "New IP address to update")
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&proxied, "proxied", true, "Whether to enable Cloudflare proxying (default: true)")
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Int16Var(&ttl, "ttl", 3600, "Time to live for the DNS record (default: 3600 seconds)")
	rootCmd.AddCommand(updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
}

func UpdateARecord(apiToken, domain, fqdn, newIP string) error {
	client := &http.Client{}

	zoneResp, err := getJson(client, fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", domain), apiToken)
	if err != nil {
		return err
	}
	var zoneResult DnsRecordsResponse
	json.Unmarshal(zoneResp, &zoneResult)

	if len(zoneResult.Result) == 0 {
		return fmt.Errorf("zone not found for %s", domain)
	}
	zoneID := zoneResult.Result[0].ID

	recordResp, err := getJson(client, fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", zoneID, fqdn), apiToken)
	if err != nil {
		return err
	}
	var recordResult DnsRecordsResponse
	json.Unmarshal(recordResp, &recordResult)
	if len(recordResult.Result) == 0 {
		return fmt.Errorf("A record not found for %s", fqdn)
	}
	recordID := recordResult.Result[0].ID

	updatePayload := map[string]interface{}{
		"type":    "A",
		"name":    fqdn,
		"content": newIP,
		"ttl":     ttl,
		"proxied": proxied,
	}
	payloadBytes, _ := json.Marshal(updatePayload)

	req, err := http.NewRequest("PUT", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+apiToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("✅ Updated %s to %s\nResponse: %s\n", fqdn, newIP, string(body))
	return nil
}

func getJson(client *http.Client, url, apiToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
