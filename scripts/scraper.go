package scripts

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
)

// Lead represents the raw data from Google Maps
type Lead struct {
	Name    string
	Phone   string
	Website string
	Rating  string
}

func ScrapeLeads(searchQuery string) ([]Lead, error) {
	// 1. Setup Stealth Context
	fmt.Println("Starting ScrapeLeads func")

	opts := GetRandomstealthOpts()

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var leads []Lead

	// 2. The Execution Loop
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.google.com/maps"),
		chromedp.WaitVisible(`form[action*="consent.google.com"] button`, chromedp.ByQuery),
		humanMove(20, 27),
		chromedp.Click(`form[action*="consent.google.com"] button:nth-child(1)`, chromedp.ByQuery),
		
		// Search for the Niche (e.g., "Shortlet Lagos")
		humanMove(55, 44),
		chromedp.WaitVisible(`#ucc-1`, chromedp.ByID),
		chromedp.SendKeys(`#ucc-1`, searchQuery+"\n", chromedp.ByID),
		chromedp.WaitVisible(`div[role="feed"], div[role="main"]`, chromedp.ByQuery),
		
		// Wait for the result list to load
		chromedp.Sleep(5 * time.Second),)
		

		if err != nil {return nil, err }

		for i := 0; i < 10; i++ {
    		fmt.Printf("Deep Scroll Cycle %d...\n", i+1)
    		err = chromedp.Run(ctx,
        		// Target the 'feed' role specifically and scroll it down 2000 pixels
        		chromedp.Evaluate(`document.querySelector('div[role="feed"]').scrollBy(0, 2000)`, nil),
        		chromedp.Sleep(5 * time.Second), // Give the network time to fetch new leads
    		)
    		if err != nil {
        		fmt.Println("Scroll error:", err)
    		}
		}

		// After your scrolling is done
		var leadNodes []*cdp.Node
		if err = chromedp.Run(ctx, chromedp.Nodes(`div[role="article"]`, &leadNodes, chromedp.ByQueryAll)); err != nil {
			return nil, err
		}

		for i, _ := range leadNodes {
    		var name, phone, website string
    		sel := fmt.Sprintf(`(//div[@role="article"])[%d]//a[@class="hfpxzc"]`, i+1) // XPath is safer here
    
    		err = chromedp.Run(ctx,
        		chromedp.Click(sel, chromedp.BySearch),
        		chromedp.Sleep(2 * time.Second), // Wait for the side panel to slide out
        		// Target the phone number specifically in the side panel
        		chromedp.Evaluate(`
            		(() => {
                		const name = document.querySelector('h1.DUwDvf')?.innerText || "";
                		const phone = document.querySelector('button[data-tooltip*="phone"], button[data-value*="Phone"]')?.innerText || "";
                		const web = document.querySelector('a[data-tooltip*="website"], a[data-value*="Website"]')?.href || "";
                		return { name, phone, web };
            		})()
        		`, &struct {
					Name	*string `json:"name"`
					Phone	*string `json:"phone"`
					Web	*string `json:"web"`
				}{&name, &phone, &website}),
    		)
    		// Save lead data here...
		}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
    		Array.from(document.querySelectorAll('div[role="article"]')).map(e => {
        		// Find the Name: usually inside an aria-label or a specific class
        		const nameLink = e.querySelector('a.hfpxzc');
        		const name = nameLink ? nameLink.getAttribute('aria-label') : "";

        		// Find the Website: Look for the specific icon-based link
        		const websiteLink = e.querySelector('a[data-value*="Website"]');
        		const website = websiteLink ? websiteLink.href : "";

        		// Find the Phone: Harder to find, usually in a specific div
        		const phoneElem = e.querySelector('span[class*="Usd1k"]'); // Or look for text with /+234/
        		const phone = phoneElem ? phoneElem.innerText : "";

        		return {
            		name: name,
            		phone: phone,
            		website: website,
            		rating: "" // We can pull this later
        		};
    		}).filter(lead => lead.name !== "") // Remove empty artifacts
	`, &leads),
		chromedp.Sleep(50* time.Second),
	)
	fmt.Println("This is the leads: ", leads)

	return leads, err
}


func humanMove(x, y int64) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var curX, curY float64 = 0, 0 // Use float64 here
		steps := 10
		targetX := float64(x)
		targetY := float64(y)

		for i := 1; i <= steps; i++ {
			stepX := curX + (targetX-curX)*float64(i)/float64(steps) + float64(time.Now().UnixNano()%10)
			stepY := curY + (targetY-curY)*float64(i)/float64(steps) + float64(time.Now().UnixNano()%10)

			// Use chromedp.MouseEvent(...).Do(ctx) inside the ActionFunc
			if err := chromedp.MouseEvent("mouseMoved", stepX, stepY).Do(ctx); err != nil {
				return err
			}
			time.Sleep(time.Duration(10+time.Now().UnixNano()%50) * time.Millisecond)
		}
		return nil
	})
}


func GetRandomstealthOpts() []chromedp.ExecAllocatorOption {
	resolutions := [][]int{
		{1920, 1080}, {1366, 768}, {1440, 900}, {1536, 864},
	}
	res := resolutions[time.Now().UnixNano()%int64(len(resolutions))]

	return append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(res[0], res[1]),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		chromedp.Flag("headless", false),
		chromedp.UserAgent(getRandomUA()), // Function to return a random Chrome UA
	)
}


func getRandomUA() string {
	uas := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
	return uas[time.Now().UnixNano()%int64(len(uas))]
}