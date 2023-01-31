package bverfg

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/goodsign/monday"
)

const (
	bverfgDomain = "bundesverfassungsgericht.de"

	senateDecisionsURL = "https://www.bundesverfassungsgericht.de/DE/Presse/Senatsbeschl%C3%BCsse/Senatsbeschl%C3%BCsse_node.html"

	upcomingDecisionAllocationSize = 5
)

// AnnouncedDecision is a court decision that's release has been announced, but
// is not published yet.
type AnnouncedDecision struct {
	Ref         CaseReference
	Description string
	// PublishDate is holds the timestamp with the announced publish date
	PublishDate time.Time
}

func GetUpcomingSenateDecisions() []AnnouncedDecision {
	upcomingDecisions := make([]AnnouncedDecision, 0, upcomingDecisionAllocationSize)

	c := colly.NewCollector(colly.AllowedDomains(bverfgDomain, fmt.Sprintf("www.%s", bverfgDomain)))

	c.OnRequest(func(r *colly.Request) {
		log.Println("scraping", r.URL.String())
	})

	c.OnHTML(`table[class="MsoNormalTable"] tr`, func(h *colly.HTMLElement) {
		// We don't want th cells possibly included here
		dataCells := h.DOM.Find("td")
		if dataCells.Length() == 0 {
			return
		}

		// Only take the first case ref if multiple (though if never seens this on this page)
		caseRefStr := dataCells.Eq(0).Text()
		description := dataCells.Eq(1).Text()
		dateStr := dataCells.Eq(2).Text()

		if caseRefStr == "" || description == "" || dateStr == "" {
			return
		}

		// Only take the first case ref if multiple (though i've never seen this on this page yet)
		caseRefStr = strings.Split(caseRefStr, ",")[0]
		caseRef, err := ParseCaseRef(caseRefStr)
		if err != nil {
			log.Printf("error parsing scraped caseref %v: %v", caseRefStr, err)
			return
		}

		// Parse german date string
		loc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			log.Panicln("error loading fixed timezone location:", err)
		}
		date, err := monday.ParseInLocation(monday.DefaultFormatDeDELong, dateStr, loc, monday.LocaleDeDE)
		if err != nil {
			log.Printf("error parsing local date str %v: %v", dateStr, err)
		}

		upcomingDecisions = append(upcomingDecisions, AnnouncedDecision{
			Ref:         caseRef,
			Description: description,
			PublishDate: date,
		})
	})

	c.Visit(senateDecisionsURL)

	return upcomingDecisions
}
