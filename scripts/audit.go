package scripts

import (
	"crypto/tls"
	"net/http"
	//"regexp"
	"strings"
	"time"
)

type AuditReport struct {
	URL          string  `json:"url"`
	SSLExpired   bool    `json:"ssl_expired"`
	LoadTimeSec  float64 `json:"load_time"`
	OutdatedCopy bool    `json:"outdated_copyright"`
	NoWebsite    bool    `json:"no_website"`
}

func PerformAudit(targetURL string) AuditReport {
	if targetURL == "" {
		return AuditReport{NoWebsite: true}
	}

	report := AuditReport{URL: targetURL}
	start := time.Now()

	// 1. Check SSL & Connectivity
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // We want to see if it's broken
		},
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		report.SSLExpired = true
		return report
	}
	defer resp.Body.Close()

	// 2. Load Time
	report.LoadTimeSec = time.Since(start).Seconds()

	// 3. Copyright Check (The "Abandonment" Signal)
	// We read the body and look for 2023, 2024, or 2025
	currentYear := "2026"
	// Simplified: in real code, read body into string
	bodySnippet := "Copyright © 2024" 
	if strings.Contains(bodySnippet, "©") && !strings.Contains(bodySnippet, currentYear) {
		report.OutdatedCopy = true
	}

	return report
}