package main

import (
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

	for i := 0; i < len(leads); i++ {
		fmt.Printf(">>> Loading... %d", i+1)
		scripts.PerformAudit(leads[i].Website) // Performing audit on the Target URL
		fmt.Printf(">>> Just Peformed Audit on this URL: %s", leads[i].Website)
	} 
}
