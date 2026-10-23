package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	const did = "did:plc:plcd6wiib6ddc2dy4t4oosvw"

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	cursor := ""
	seen := make(map[string]bool)
	count := 0

	for {
		params := url.Values{}
		params.Set("actor", did)
		params.Set("filter", "posts_no_replies")
		params.Set("limit", "100")

		if cursor != "" {
			params.Set("cursor", cursor)
		}

		endpoint := "https://public.api.bsky.app/xrpc/app.bsky.feed.getAuthorFeed?" + params.Encode()

		resp, err := client.Get(endpoint)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Fatalf("Bluesky API: %s", resp.Status)
		}

		var data struct {
			Cursor string `json:"cursor"`
			Feed   []struct {
				Reason *struct{} `json:"reason"`
				Post   Post      `json:"post"`
			} `json:"feed"`
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()

		if err != nil {
			log.Fatal(err)
		}

		for _, item := range data.Feed {
			p := item.Post

			if item.Reason != nil ||
				p.Author.DID != did ||
				p.Record.Reply != nil ||
				seen[p.URI] {
				continue
			}

			seen[p.URI] = true
			generate(p.URI)
			count++
		}

		if data.Cursor == "" ||
			data.Cursor == cursor ||
			len(data.Feed) == 0 {
			break
		}

		cursor = data.Cursor
	}

	fmt.Printf("Обработано тредов: %d\n", count)
}
