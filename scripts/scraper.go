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

		for i := 0; i < 2; i++ { // use i < 10
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
		var seen = make(map[string]bool)
		var leadNodes []*cdp.Node
		if err = chromedp.Run(ctx, chromedp.Nodes(`div[role="article"]`, &leadNodes, chromedp.ByQueryAll)); err != nil {
			return nil, err
		}

		for i, _ := range leadNodes {
    		var name, phone, website string
			containerSel := fmt.Sprintf(`(//div[@role="article"])[%d]`, i+1)
			linkSel := containerSel + `//a[@class="hfpxzc"]`

			fmt.Printf("Analyzing Lead %d/%d...\n", i+1, len(leadNodes))
    
    		err = chromedp.Run(ctx,
				chromedp.ScrollIntoView(containerSel, chromedp.BySearch),
				chromedp.Sleep(500 * time.Millisecond),
        		chromedp.Click(linkSel, chromedp.BySearch),
        		chromedp.Sleep(10000 * time.Millisecond), // Wait for the side panel to slide out
        		// Target the phone number specifically in the side panel
        		chromedp.Evaluate(`
            		(() => {
					 	const panelName = document.querySelector('h1.DUwDvf')?.innerText || "";	
                		const phone = document.querySelector('button[data-tooltip*="phone"], button[data-value*="Phone"]')?.innerText || "";
						const websiteElem = document.querySelector('a.CsEnBe[data-tooltip*="website"], a.CsEnBe[data-value*="Website"]');
        				let web = "";

						if (websiteElem){
							web = websiteElem.href;
						}
                		return { name: panelName, phone, web };
            		})()
        		`, &struct {
					Name	*string `json:"name"`
					Phone	*string `json:"phone"`
					Web		*string `json:"web"`
				}{&name, &phone, &website}),
    		)
    		// Save lead data here...
			if err == nil && name != "" {
				fingerprint := phone
				if phone == "" {
					fingerprint = name
				}
				if !seen[fingerprint] {
					seen[fingerprint] = true
					leads = append(leads, Lead{
						Name: name,
						Phone: phone,
						Website: website,
					})
					fmt.Printf("Successfully Captured: %s\n", name)
					fmt.Printf("Captured - Name: %s | Phone: %s | Web: %s\n", name, phone, website)
				}else{
					fmt.Printf(">>>Skipping Duplicate... %s\n", fingerprint)
				}
			}
		}

	fmt.Printf(">>> This is the leads: %s", leads)
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
		chromedp.Flag("headless", true), // Change to false to see the bot scrape in real-time
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.UserAgent(getRandomUA()), // Function to return a random Chrome UA
	)
}

func getRandomUA() string {
    uas := []string{
        // Windows Chrome/Edge
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/121.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
        // Mac Safari/Chrome
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
        // Linux Firefox
        "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
        "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0",
        // Mobile (Crucial for "stealth" as mobile UAs often get less scrutiny)
        "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
        "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1",
        "Mozilla/5.0 (Linux; Android 14; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6312.40 Mobile Safari/537.36",
    }
    return uas[time.Now().UnixNano()%int64(len(uas))]
}
