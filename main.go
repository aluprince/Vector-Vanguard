package main

import (
	"encoding/json"
	"os"
	"github.com/aluprince/Vector-Vanguard/scripts"
	"fmt"
	"log"
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
		fmt.Printf(">>> Just Peformed Audit on this URL: %s", leads[i].Website)
		saveReport(allReports)
	} 
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

// Inside main.go
// resp, _ := http.Post("http://localhost:5000/new_leads", "application/json", bytes.NewBuffer(jsonData))