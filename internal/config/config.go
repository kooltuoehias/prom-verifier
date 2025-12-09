package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the runtime configuration for the verifier.
type Config struct {
	FilePath     string
	PromURL      string
	OutputFormat string // Added
	Start        time.Time
	End          time.Time
	RuleFile     RuleFile
}

// Load parses command line flags and loads the configuration.
// It returns a validated Config or an error.
func Load() (*Config, error) {
	// Define flags
	filePath := flag.String("file", "alert.yaml", "Path to the alert rule file (YAML/TF)")
	promURL := flag.String("url", "http://localhost:9090", "Prometheus API URL")
	atTimestamp := flag.String("at", "", "Target timestamp (e.g., '2023-11-20 14:30'). Default is Now.")
	windowStr := flag.String("window", "30m", "Time window around the target (e.g., 30m means target +/- 30m)")
	outputFormat := flag.String("output", "text", "Output format (text, json, yaml)") // Added

	flag.Parse()

	// Parse window duration
	window, err := time.ParseDuration(*windowStr)
	if err != nil {
		return nil, fmt.Errorf("invalid window format: %w", err)
	}
	// Safety: 4 hour window limit to prevent massive queries
	if window > 4*time.Hour {
		return nil, fmt.Errorf("safety block: window size %s exceeds maximum allowed limit of 4h. "+
			"please choose a smaller window to avoid crashing prometheus", window)
	}

	// Calculate Target Time
	var targetTime time.Time
	if *atTimestamp == "" {
		targetTime = time.Now()
		fmt.Fprintln(os.Stderr, "🕒 Mode: Realtime (Now)")
	} else {
		layout := "2006-01-02 15:04"
		targetTime, err = time.ParseInLocation(layout, *atTimestamp, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid time format, please use 'YYYY-MM-DD HH:MM': %w", err)
		}
		fmt.Fprintf(os.Stderr, "🕰️ Mode: Time Travel (Target: %s)\n", targetTime.Format(layout))
	}

	// Calculate Start and End
	start := targetTime.Add(-window)
	end := targetTime.Add(window)

	// Validate safety (Max 3 months old)
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	if start.Before(threeMonthsAgo) {
		return nil, fmt.Errorf("safety block: query start time %s is older than 3 months limit (%s)",
			start.Format("2006-01-02"), threeMonthsAgo.Format("2006-01-02"))
	}

	fmt.Fprintf(os.Stderr, "🔍 Replay Window: %s  <---->  %s\n", start.Format("15:04"), end.Format("15:04"))
	fmt.Fprintln(os.Stderr, "---------------------------------------------------")
	fmt.Fprintln(os.Stderr, "🔧 Configuration:")
	fmt.Fprintf(os.Stderr, "   File:  %s\n", *filePath)
	fmt.Fprintf(os.Stderr, "   URL:   %s\n", *promURL)
	fmt.Fprintln(os.Stderr, "---------------------------------------------------")

	// Read and Parse Rule File
	data, err := os.ReadFile(*filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", *filePath, err)
	}

	var ruleFile RuleFile
	if err := yaml.Unmarshal(data, &ruleFile); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return &Config{
		FilePath:     *filePath,
		PromURL:      *promURL,
		OutputFormat: *outputFormat, // Added
		Start:        start,
		End:          end,
		RuleFile:     ruleFile,
	}, nil
}
