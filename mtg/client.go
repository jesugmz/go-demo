package mtg

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

// Client provides MTG data retrieval operations.
type Client interface {
	Fetch(page int) ([]Card, error)
	FetchWithMetaData() ([]Card, map[string]int, error)
}

const defaultBaseEndpoint = "https://api.magicthegathering.io/v1"

type client struct {
	baseEndpoint string
}

// NewThrottler creates a Client service with the given dependencies.
func NewClient() Client {
	// The reason why we are actually returning a "client" rather than "Client"
	// data type is because "Client" is the interface - an empty data type in
	// Golang - which can be satisfied by the implementation of any data type.
	return &client{baseEndpoint: defaultBaseEndpoint}
}

// Fetch fetches the given page from the MTG API.
func (c *client) Fetch(page int) ([]Card, error) {
	resp, err := c.fetch(page)
	if err != nil {
		return []Card{}, err
	}
	// resp.Body is a data stream that must be closed once we finish with.
	defer resp.Body.Close()

	return Decode(resp.Body)
}

// FetchWithMetaData fetches data and headers for the first page from the MTG API.
func (c *client) FetchWithMetaData() ([]Card, map[string]int, error) {
	metaData := map[string]int{}
	resp, err := c.fetch(1)
	if err != nil {
		return []Card{}, metaData, err
	}
	// resp.Body is a data stream that must be closed once we finish with.
	defer resp.Body.Close()

	cards, err := Decode(resp.Body)
	re := regexp.MustCompile(`(?:.*page=)(?P<pages>\d+)(?:>;\srel="last")`)
	// Access to the index 1 because the first contains the full match.
	metaData["totalPages"], err = strconv.Atoi(re.FindStringSubmatch(resp.Header.Get("Link"))[1])
	if err != nil {
		return []Card{}, metaData, errors.New(fmt.Sprintf("could not extract Total Pages from headers: %v", err))
	}
	metaData["rateLimit"], err = strconv.Atoi(resp.Header.Get("Ratelimit-Remaining"))
	if err != nil {
		return []Card{}, metaData, errors.New(fmt.Sprintf("could not extract Rate Limit from headers: %v", err))
	}

	return cards, metaData, err
}

func (c *client) fetch(page int) (*http.Response, error) {
	// A typical error is to trust in the server health so we can be blocked
	// for a long time or even indefinitely.
	// As MTG API is returning me very different timeout values I just explain
	// it here to speed up.
	resp, err := http.Get(c.baseEndpoint + "/cards?page=" + strconv.Itoa(page))
	if err != nil {
		return &http.Response{}, errors.New(fmt.Sprintf("could not fetch data: %v", err))
	}

	err = c.responseError(resp)
	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func (c *client) responseError(resp *http.Response) error {
	if resp.StatusCode == 200 {
		return nil
	}

	// TODO: define the proper struct
	var serverErr string
	if err := json.NewDecoder(resp.Body).Decode(&serverErr); err != nil {
		return errors.New(fmt.Sprintf("could not decode response error: %v", resp.Body))
	}

	return errors.New(serverErr)
}
