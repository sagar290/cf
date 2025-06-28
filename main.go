package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
var proxied bool
var ttl int16 = 3600
var dnsRecord string = "A"

func main() {
	rootCmd := &cobra.Command{Use: "cf"}

	updateCmd := &cobra.Command{
		Use:   "update:dns [domain] [type] [key] [value]",
		Short: "update:dns specific A record (like root or www) for a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			dnsRecord := args[1]
			key := args[2]
			value := args[3]

			if key == "" || value == "" {
				return fmt.Errorf("key and value must be provided")
			}

			if apiToken == "" {
				apiToken = os.Getenv("CF_API_TOKEN")
			}

			if apiToken == "" {
				return fmt.Errorf("cloudflare API token not provided")
			}

			return UpdateDNSRecord(apiToken, domain, dnsRecord, key, value)
			// return nil
		},
	}

	updateCmd.Flags().BoolVar(&proxied, "proxied", true, "Whether to enable Cloudflare proxying (default: true)")
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Int16Var(&ttl, "ttl", 3600, "Time to live for the DNS record (default: 3600 seconds)")
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVar(&dnsRecord, "dnsRecord", "A", "Type of DNS record to update (default: A)")
	rootCmd.AddCommand(updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
}

func UpdateDNSRecord(apiToken, domain string, dnsRecord string, key string, value string) error {
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

	if key == "@" {
		key = domain
	}

	fqdn := key
	if key != "" && !keyHasDomainSuffix(key, domain) {
		fqdn = key + "." + domain
	}

	recordResp, err := getJson(client, fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", zoneID, fqdn), apiToken)

	fmt.Println("🔍 Searching for A record:", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", zoneID, fqdn), fqdn, string(recordResp))

	if err != nil {
		return err
	}
	var recordResult DnsRecordsResponse
	err = json.Unmarshal(recordResp, &recordResult)
	if err != nil {
		return err
	}
	if len(recordResult.Result) == 0 {
		return fmt.Errorf("the A record not found for %s", fqdn)
	}
	recordID := recordResult.Result[0].ID

	updatePayload := map[string]interface{}{
		"type":    dnsRecord,
		"name":    key,
		"content": value,
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("✅ Updated %s to %s\nResponse: %s\n", key, value, string(body))
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

func keyHasDomainSuffix(key string, domain string) bool {
	if len(key) < len(domain) {
		return false
	}
	return key[len(key)-len(domain):] == domain
}
