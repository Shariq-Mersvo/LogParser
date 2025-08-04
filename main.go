package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// PostRequest represents a POST request found in the logs
type PostRequest struct {
	Timestamp time.Time `json:"timestamp"`
	Endpoint  string    `json:"endpoint"`
	Duration  string    `json:"duration"`
	IP        string    `json:"ip"`
	Date      string    `json:"date"`
	Time      string    `json:"time"`
}

// UsageStats represents usage statistics for analysis
type UsageStats struct {
	TotalPOSTRequests int                       `json:"total_post_requests"`
	EndpointCounts    map[string]int            `json:"endpoint_counts"`
	DailyUsage        map[string]int            `json:"daily_usage"`
	HourlyUsage       map[string]int            `json:"hourly_usage"`
	EndpointsByDay    map[string]map[string]int `json:"endpoints_by_day"`
}

// getEndpointAlias returns a user-friendly name for endpoints
func getEndpointAlias(endpoint string) string {
	aliases := map[string]string{
		"/modes":         "Room Mode Changes",
		"/lutron/shades": "Shade Controls",
		"/iptv/channel":  "TV Controls",
		"/iptv/remote":   "TV Controls",
		"/iptv":          "TV Controls",
		"/bacnet/info":   "AC Temperature",
		"/cyviz/avinput": "Cyviz TV Controls",
	}

	if alias, exists := aliases[endpoint]; exists {
		return alias
	}
	return endpoint // fallback to original if no alias found
}

func main() {
	logsDir := "logs"

	// Check if logs directory exists
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		log.Fatalf("Logs directory '%s' does not exist", logsDir)
	}

	// Find all log files
	logFiles, err := filepath.Glob(filepath.Join(logsDir, "server_*.log"))
	if err != nil {
		log.Fatalf("Error finding log files: %v", err)
	}

	if len(logFiles) == 0 {
		log.Fatalf("No log files found in %s directory", logsDir)
	}

	fmt.Printf("Found %d log files\n", len(logFiles))

	// Initialize stats
	stats := &UsageStats{
		EndpointCounts: make(map[string]int),
		DailyUsage:     make(map[string]int),
		HourlyUsage:    make(map[string]int),
		EndpointsByDay: make(map[string]map[string]int),
	}

	// Regular expression to match GIN POST requests
	// [GIN] 2025/05/19 - 23:24:39 | 200 |    114.1859ms |             ::1 | POST     "/modes"
	ginRegex := regexp.MustCompile(`\[GIN\]\s+(\d{4}/\d{2}/\d{2})\s+-\s+(\d{2}:\d{2}:\d{2})\s+\|\s+\d+\s+\|\s+([^\|]+)\s+\|\s+([^\|]+)\s+\|\s+POST\s+\"([^\"]+)\"`)

	// Process each log file
	for _, logFile := range logFiles {
		fmt.Printf("Processing: %s\n", logFile)

		file, err := os.Open(logFile)
		if err != nil {
			log.Printf("Error opening file %s: %v", logFile, err)
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		fileRequestCount := 0

		for scanner.Scan() {
			line := scanner.Text()

			// Check if line contains a POST request
			if strings.Contains(line, "POST") && strings.Contains(line, "[GIN]") {
				matches := ginRegex.FindStringSubmatch(line)
				if len(matches) == 6 {
					dateStr := matches[1]  // 2025/05/19
					timeStr := matches[2]  // 23:24:39
					endpoint := matches[5] // /modes

					// Update statistics
					stats.TotalPOSTRequests++

					// Use aliases for endpoint counting
					endpointAlias := getEndpointAlias(endpoint)
					stats.EndpointCounts[endpointAlias]++
					stats.DailyUsage[dateStr]++

					// For hourly usage, just use the hour (00-23) to aggregate across all days
					hour := timeStr[:2] // Extract hour from "23:24:39"
					stats.HourlyUsage[hour+":00"]++

					// Update endpoints by day (using aliases)
					if stats.EndpointsByDay[dateStr] == nil {
						stats.EndpointsByDay[dateStr] = make(map[string]int)
					}
					stats.EndpointsByDay[dateStr][endpointAlias]++
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading file %s: %v", logFile, err)
		}

		fmt.Printf("  Found %d POST requests\n", fileRequestCount)
	}

	// Generate reports
	generateTextReport(stats)
	generateJSONReport(stats)

	fmt.Printf("\nAnalysis complete!\n")
	fmt.Printf("Total POST requests found: %d\n", stats.TotalPOSTRequests)
	fmt.Printf("Reports saved: usage_stats.txt and usage_stats.json\n")
}

func generateTextReport(stats *UsageStats) {
	file, err := os.Create("usage_stats.txt")
	if err != nil {
		log.Printf("Error creating text report: %v", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "SERVER USAGE STATISTICS REPORT\n")
	fmt.Fprintf(file, "===============================\n\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Fprintf(file, "OVERVIEW\n")
	fmt.Fprintf(file, "--------\n")
	fmt.Fprintf(file, "Total POST Requests: %d\n\n", stats.TotalPOSTRequests)

	// Endpoint statistics
	fmt.Fprintf(file, "ENDPOINT USAGE\n")
	fmt.Fprintf(file, "--------------\n")

	// Sort endpoints by usage count
	type endpointCount struct {
		endpoint string
		count    int
	}
	var endpoints []endpointCount
	for endpoint, count := range stats.EndpointCounts {
		endpoints = append(endpoints, endpointCount{endpoint, count})
	}
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].count > endpoints[j].count
	})

	for _, ep := range endpoints {
		percentage := float64(ep.count) / float64(stats.TotalPOSTRequests) * 100
		fmt.Fprintf(file, "%-20s: %5d requests (%.1f%%)\n", ep.endpoint, ep.count, percentage)
	}

	// Daily usage
	fmt.Fprintf(file, "\nDAILY USAGE\n")
	fmt.Fprintf(file, "-----------\n")

	// Sort dates
	var dates []string
	for date := range stats.DailyUsage {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	for _, date := range dates {
		count := stats.DailyUsage[date]
		fmt.Fprintf(file, "%s: %d requests\n", date, count)
	}

	// Hourly usage (aggregated across all days)
	fmt.Fprintf(file, "\nHOURLY USAGE (AGGREGATED)\n")
	fmt.Fprintf(file, "-------------------------\n")

	// Sort hours
	var hours []string
	for hour := range stats.HourlyUsage {
		hours = append(hours, hour)
	}
	sort.Strings(hours)

	for _, hour := range hours {
		count := stats.HourlyUsage[hour]
		fmt.Fprintf(file, "%s: %d requests\n", hour, count)
	}

	// Detailed daily breakdown by endpoint
	fmt.Fprintf(file, "\nDETAILED DAILY BREAKDOWN\n")
	fmt.Fprintf(file, "------------------------\n")
	for _, date := range dates {
		if endpointMap, exists := stats.EndpointsByDay[date]; exists {
			fmt.Fprintf(file, "\n%s:\n", date)

			// Sort endpoints for this day
			var dayEndpoints []endpointCount
			for endpoint, count := range endpointMap {
				dayEndpoints = append(dayEndpoints, endpointCount{endpoint, count})
			}
			sort.Slice(dayEndpoints, func(i, j int) bool {
				return dayEndpoints[i].count > dayEndpoints[j].count
			})

			for _, ep := range dayEndpoints {
				fmt.Fprintf(file, "  %-18s: %d requests\n", ep.endpoint, ep.count)
			}
		}
	}
}

func generateJSONReport(stats *UsageStats) {
	file, err := os.Create("usage_stats.json")
	if err != nil {
		log.Printf("Error creating JSON report: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(stats); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
