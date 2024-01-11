package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	var err error

	// Command-line flags
	pushgatewayUrl := "http://localhost:9091" // Default Pushgateway URL
	jobName := flag.String("j", "", "The job name")
	metricName := flag.String("n", "", "The name of the metric")
	metricType := flag.String("t", "gauge", "The type of the metric (gauge, counter, etc.)")
	metricHelp := flag.String("h", "", "Help text for the metric")
	cmdLineLabels := flag.String("l", "", "Comma-separated labels in key=value format")

	flag.Parse()

	if *jobName == "" {
		fmt.Fprintln(os.Stderr, "Error: job name is required")
		// Display usage information
		flag.Usage()
		os.Exit(1)
	}

	if *metricName == "" {
		fmt.Fprintln(os.Stderr, "Error: metric name is required")
		// Display usage information
		flag.Usage()
		os.Exit(1)
	}

	// Combine labels from environment variable and command-line flag
	var labels string
	if labels, err = parseLabels("PROMPIPE_LABELS", *cmdLineLabels); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing labels:", err)
		os.Exit(1)
	}

	// Read metric value from stdin
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Fprintln(os.Stderr, "Error reading metric value from stdin")
		os.Exit(1)
	}
	metricValue := scanner.Text()

	// Format the data in Prometheus exposition format
	var dataBuilder strings.Builder
	if *metricHelp != "" {
		dataBuilder.WriteString(fmt.Sprintf("# HELP %s %s\n", *metricName, *metricHelp))
	}
	dataBuilder.WriteString(fmt.Sprintf("# TYPE %s %s\n", *metricName, *metricType))
	metricLine := fmt.Sprintf("%s{%s} %s\n", *metricName, labels, metricValue)
	dataBuilder.WriteString(metricLine)

	// Send the data to the Pushgateway
	url := fmt.Sprintf("%s/metrics/job/%s", pushgatewayUrl, *jobName)

	resp, err := http.Post(url, "text/plain", strings.NewReader(dataBuilder.String()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error sending request to Pushgateway:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the body of the response
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to push to Pushgateway: %s: %s\n", resp.Status, body)
		os.Exit(1)
	}

	fmt.Println("Metric pushed to Pushgateway successfully:", metricLine)
}

// formatLabel formats label as key="value" (including the double quotes)
func formatLabel(raw string) (string, error) {
	fmt.Println("formatLabel", raw)
	parts := strings.Split(raw, "=")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid label: %s", raw)
	}
	if strings.Contains(parts[0], "\"") {
		return "", fmt.Errorf("invalid label: %s", raw)
	}
	if strings.Contains(parts[1], "\"") {
		if strings.HasPrefix(parts[1], "\"") && strings.HasSuffix(parts[1], "\"") {
			return fmt.Sprintf(`%s=%s`, parts[0], parts[1]), nil
		} else {
			return "", fmt.Errorf("invalid label: %s", raw)
		}
	} else {
		return fmt.Sprintf(`%s="%s"`, parts[0], parts[1]), nil
	}
}

func formatLabels(raw string) (string, error) {
	var fmtLabels []string
	for _, label := range strings.Split(raw, ",") {
		fmtLabel, err := formatLabel(label)
		if err != nil {
			return "", err
		}
		fmtLabels = append(fmtLabels, fmtLabel)
	}
	return strings.Join(fmtLabels, ","), nil
}

func parseLabels(envVar string, cmdLineLabels string) (string, error) {
	envLabels := os.Getenv(envVar)
	var allLabels string
	if envLabels != "" && cmdLineLabels != "" {
		allLabels = envLabels + "," + cmdLineLabels
	} else if envLabels != "" {
		allLabels = envLabels
	} else {
		allLabels = cmdLineLabels
	}
	if allLabels == "" {
		return "", nil
	}
	return formatLabels(allLabels)
}
