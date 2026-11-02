package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type MariaDBV2Cluster struct {
	ID      string
	Name    string
	Version string
}

type MariaDBV2ItemProperties struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MariaDBV2Item struct {
	ID         string                     `json:"id"`
	Properties MariaDBV2ItemProperties `json:"properties"`
}

type MariaDBV2Collection struct {
	Items []MariaDBV2Item `json:"items"`
}

var (
	mariadbV2Clusters   []MariaDBV2Cluster
	mariadbV2FetchError float64
	mariadbV2Mutex      sync.RWMutex
)

func MariaDBV2CollectResources(cycletime int32) {
	for {
		scrapeMariaDBV2()
		time.Sleep(time.Duration(cycletime) * time.Second)
	}
}

func scrapeMariaDBV2() {
	token := os.Getenv("IONOS_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "TOKEN environment variable is not set")
		os.Exit(1)
	}
  
	endpoints := []string{
		"https://mariadb.de-txl.ionos.com/v2/clusters",
		"https://mariadb.de-fra.ionos.com/v2/clusters",
	}

	clusters := make([]MariaDBV2Cluster, 0)
	fetchError := 0.0

	for _, endpoint := range endpoints {
		collection, err := fetchMariaDBV2Clusters(token, endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", endpoint, err)
			fetchError = 1.0
			continue
		}
		for idx := range collection.Items {
			item := collection.Items[idx]

			clusterID := "UNKNOWN"
			if item.ID != "" {
				clusterID = item.ID
			}
			clusterName := "UNKNOWN"
			if item.Properties.Name != "" {
				clusterName = item.Properties.Name
			}
			clusterVersion := "UNKNOWN"
			if item.Properties.Version != "" {
				clusterVersion = item.Properties.Version
			}

			clusters = append(clusters, MariaDBV2Cluster{
				ID:      clusterID,
				Name:    clusterName,
				Version: clusterVersion,
			})
		}
	}

	mariadbV2Mutex.Lock()
	mariadbV2Clusters = clusters
	mariadbV2FetchError = fetchError
	mariadbV2Mutex.Unlock()
}

func fetchMariaDBV2Clusters(token, url string) (*MariaDBV2Collection, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	var collection MariaDBV2Collection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, err
	}

	return &collection, nil
}
