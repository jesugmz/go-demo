package mtg

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Client provides MTG data retrieval operations.
type Throttler interface {
	Run()
}

// We have to limit the maximum amount of simultaneous concurrency due server
// limitations. Spawning "Ratelimit-Remaining" threads it get crashed.
// As vaguely defined:
// "Third-party applications are currently throttled to 5000 requests per hour."
// 5000 requests / 1 hour <==> ~1 requests / 1 second but this seems to be too
// much conservative value for a single client so we set a higher value for this
// challenge assuming its risks.
// https://docs.magicthegathering.io/#documentationrate_limits
const maxConcurrencyPerSecond = 3

type throttler struct {
	// Maximum amount of concurrent requests we want to burst/process at a time.
	maxBurstRequests int
	// Holds the un-processed pages in order to put it back in case we need to
	// retry a page request.
	unprocessedPagesQueue chan int
	// Amount of pages that has been processed.
	processedPages        uint64
	totalPages            uint64
	client                Client
	storage               *CardStorage
}

// NewThrottler creates a Throttler service with the given dependencies.
func NewThrottler(c Client, s *CardStorage) Throttler {
	return &throttler{
		maxBurstRequests:      maxConcurrencyPerSecond,
		unprocessedPagesQueue: make(chan int),
		processedPages:        0,
		totalPages:            0,
		client:                c,
		storage:               s,
	}
}

// Run runs the throttler by, firstly, provisioning the needed resources.
func (t *throttler) Run() {
	defer close(t.unprocessedPagesQueue)

	// MTG API does not supports HEAD requests in order to retrieve meta data
	// like number of pages so we could foresee and reserve the exact amount of
	// resources we need in a reliable way. This means we have the following
	// possible scenarios:
	// 1. block previous requests to read the next pages from
	// 2. make an initial and separated request to gather all the current meta-data from
	//
	// Option 2 is faster in performance and easier to implement with the
	// inconvenient of loosing a possible new page after the initial meta-data
	// gathering.
	//
	// For the sake of simplicity option 2 is the chosen one.
	log.Print("Gathering meta data...")
	t.gatherMetaData()

	var wg sync.WaitGroup
	// Wait until all the work has been finished before returning.
	defer wg.Wait()
	// The timer is going to define when we want to repeat a bursting of requests.
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	// Loop until each page has been processed.
	for {
		if t.noMorePageToProcess() {
			log.Print("All pages has been processed")
			break
		}

		// Limit our requests bursting by the defined time.
		<-timer.C

		// Bursts the maximum amount of requests at a time.
		n := t.burstRequests()
		for i := 0; i < n; i++ {
			wg.Add(1)
			go t.processPage(&wg, <-t.unprocessedPagesQueue)
		}

		// Here could be implemented a dynamic limiter. A fast idea:
		// 1. Get the "Ratelimit-Remaining" header
		// 2. If "Ratelimit-Remaining" == 0 hold for a while
		//
		// Not implemented for simplicity - no much sense for the scope of this
		// challenge.
	}
}

func (t *throttler) gatherMetaData() {
	cards, metaData, err := t.client.FetchWithMetaData()
	if err != nil {
		// Fatal - exit call - as we need this data to start working.
		log.Fatalf("%v", err)
	}
	t.unprocessedPagesQueue = make(chan int, metaData["totalPages"])
	// Here there is no race condition that forces us to use atomic counter.
	t.processedPages = 1
	t.totalPages = uint64(metaData["totalPages"])
	t.storage.Cards = append(t.storage.Cards, cards...)

	// Fill up the pages queue. It is gonna help to determine when a page has been
	// processed or need to be retried.
	for i := 2; i <= metaData["totalPages"]; i++ {
		t.unprocessedPagesQueue <- i
	}
}

func (t *throttler) noMorePageToProcess() bool {
	return atomic.LoadUint64(&t.processedPages) == t.totalPages
}

func (t *throttler) burstRequests() int {
	n := len(t.unprocessedPagesQueue)
	if n > t.maxBurstRequests {
		return t.maxBurstRequests
	}
	return n
}

func (t *throttler) processPage(wg *sync.WaitGroup, page int) {
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered from panic. Releasing resources...")
			// Put back the page to be processed again.
			t.unprocessedPagesQueue <- page
		}
	}()
	defer wg.Done()
	log.Printf("Processing page %d", page)

	cards, err := t.client.Fetch(page)
	if err != nil {
		log.Printf("%v", err)
		// Error fetching the page. Put back the page to be processed again.
		t.unprocessedPagesQueue <- page
		return
	}
	t.storage.Lock()
	t.storage.Cards = append(t.storage.Cards, cards...)
	t.storage.Unlock()

	// Safe mechanism to increment our counter under possibles race conditions.
	atomic.AddUint64(&t.processedPages, 1)
}
