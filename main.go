package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aluprince/Vector-Vanguard/scripts"
)

func main() {
	fmt.Println("Starting Scraper...")
	leads, err := scripts.ScrapeLeads("Shortlet Lagos")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d leads!\n", len(leads))

	var allReports []scripts.AuditReport

	for i := 0; i < len(leads); i++ {
		fmt.Printf(">>> Loading... %d", i+1)
		report := scripts.PerformAudit(leads[i].Name, leads[i].Phone, leads[i].Website) // Performing audit on the Target URL
		allReports = append(allReports, report)
		fmt.Printf(">> This is your report: %+v\n", report)
		// fmt.Printf(">>> Just Peformed Audit on this URL: %s", leads[i].Website)
		saveReport(allReports)
	}
	fmt.Printf("----------------------------------------------------\n")
	fmt.Printf(">>> Pushing to Bot........\n")
	pushToBot(allReports)
}

// In your main.go
func saveReport(reports []scripts.AuditReport) {
	file, err := os.Create("final_audit.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // This makes it "Pretty Printed"
	if err := encoder.Encode(reports); err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
}

func pushToBot(reports []scripts.AuditReport) {
    jsonData, err := json.Marshal(reports)
    if err != nil {
        log.Printf("Error marshalling: %v", err)
        return
    }

    resp, err := http.Post("http://localhost:5000/new_leads", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Failed to push to bot: %v. Is the Python bot running?", err)
        return
    }
    defer resp.Body.Close()
    fmt.Println("Successfully pushed to bot. Check your Telegram!")
}