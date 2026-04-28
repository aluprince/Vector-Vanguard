package scripts

import (
	"crypto/tls"
	"fmt"
	"io"         // Need this to read the body
	"net"        // REQUIRED for net.OpError
	"net/http"
	"strings"
	"time"
)

type AuditReport struct {
	URL              string  `json:"url"`
	SSLExpired       bool    `json:"ssl_expired"`
	LoadTimeSec      float64 `json:"load_time"`
	OutdatedCopy     bool    `json:"outdated_copyright"`
	NoWebsite        bool    `json:"no_website"`
	FailureReason    string  `json:"failure_reason"`
	IsBroken         bool    `json:"is_broken"`
	StatusCode       int     `json:"status_code"`
}

func PerformAudit(targetURL string) AuditReport {
	if targetURL == "" {
		return AuditReport{NoWebsite: true, FailureReason: "No URL Provided"}
	}

	report := AuditReport{URL: targetURL}
	start := time.Now()

	// 1. Setup Client
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			// We keep this false initially to detect SSL issues
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	resp, err := client.Get(targetURL)
	
	// Handle Connection Errors
	if err != nil {
		report.IsBroken = true
		if strings.Contains(err.Error(), "x509") || strings.Contains(err.Error(), "certificate") {
			report.SSLExpired = true
			report.FailureReason = "SSL/Certificate Error"
		} else if dnsErr, ok := err.(*net.OpError); ok {
			if strings.Contains(dnsErr.Error(), "no such host") {
				report.FailureReason = "Domain Expired/DNS Failure"
			} else {
				report.FailureReason = "Server Down/Connection Refused"
			}
		} else {
			report.FailureReason = "Timeout/Network Error"
		}
		return report
	}
	defer resp.Body.Close()

	// 2. Load Time Calculation
	report.LoadTimeSec = time.Since(start).Seconds()
	report.StatusCode = resp.StatusCode

	// 3. Read Body for Copyright Check
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodySnippet := string(bodyBytes)
	
	currentYear := "2026"
	// Check if page contains copyright but not the current year
	if strings.Contains(bodySnippet, "©") || strings.Contains(bodySnippet, "Copyright") {
		if !strings.Contains(bodySnippet, currentYear) {
			report.OutdatedCopy = true
		}
	}

	if resp.StatusCode >= 400 {
		report.IsBroken = true
		report.FailureReason = fmt.Sprintf("HTTP %d Error", resp.StatusCode)
	}

	return report
}