package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const postURI = "at://did:plc:plcd6wiib6ddc2dy4t4oosvw/app.bsky.feed.post/3myk2gm76rc2y"

type Post struct {
	URI    string `json:"uri"`
	Author struct {
		DID string `json:"did"`
	} `json:"author"`
	Record struct {
		Text      string `json:"text"`
		CreatedAt string `json:"createdAt"`
		Reply     *struct {
			Parent struct {
				URI string `json:"uri"`
			} `json:"parent"`
		} `json:"reply"`
	} `json:"record"`
}

type Thread struct {
	Post    Post     `json:"post"`
	Replies []Thread `json:"replies"`
}

type Page struct {
	Title      string
	Date       string
	ISODate    string
	Canonical  string
	Paragraphs []string
}

const htmlPage = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} — rpne.net</title>
<meta name="description" content="Личные заметки на rpne.net">
<link rel="canonical" href="{{.Canonical}}">
<style>
  body {
    max-width: 700px;
    margin: 0 auto;
    padding: 40px 20px;
    background: #fafafa;
    color: #222;
    font: 18px/1.7 system-ui, sans-serif;
  }
  header {
    font-size: 15px;
  }
  header {
    margin-bottom: 60px;
  }
  h1 {
    font-size: 2rem;
    line-height: 1.3;
  }
  time {
    color: #777;
    font-size: 14px;
  }
  article p {
    white-space: pre-line;
    margin: 1.4em 0;
  }
  a {
    color: #1769aa;
  }
</style>
</head>
<body>
<header>
  <a href="/">rpne.net</a>
</header>
<main>
<article>
  <h1>{{.Title}}</h1>
  <time datetime="{{.ISODate}}">{{.Date}}</time>
  {{range .Paragraphs}}
  <p>{{.}}</p>
  {{end}}
</article>
</main>
</body>
</html>
`

func paragraphs(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	var result []string
	for _, p := range strings.Split(text, "\n\n") {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func main() {
	params := url.Values{}
	params.Set("uri", postURI)
	params.Set("depth", "50")

	endpoint := "https://public.api.bsky.app/xrpc/app.bsky.feed.getPostThread?" + params.Encode()

	client := &http.Client{Timeout: 20 * time.Second}

	resp, err := client.Get(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Bluesky API: HTTP %d", resp.StatusCode)
	}

	var data struct {
		Thread Thread `json:"thread"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal(err)
	}

	root := data.Thread.Post
	if root.Author.DID == "" {
		log.Fatal("Публикация не найдена")
	}

	first := strings.SplitN(strings.TrimSpace(root.Record.Text), "\n", 2)
	title := strings.TrimSpace(first[0])
	if title == "" {
		log.Fatal("Отсутствует заголовок")
	}

	var body []string
	if len(first) > 1 {
		body = append(body, paragraphs(first[1])...)
	}

	current := &data.Thread
	parts := 1

	for {
		var next *Thread

		for i := range current.Replies {
			reply := &current.Replies[i]
			if reply.Post.Author.DID != root.Author.DID ||
				reply.Post.Record.Reply == nil {
				continue
			}
			if reply.Post.Record.Reply.Parent.URI == current.Post.URI {
				next = reply
				break
			}
		}

		if next == nil {
			break
		}

		body = append(body, paragraphs(next.Post.Record.Text)...)
		parts++
		current = next
	}

	created, err := time.Parse(time.RFC3339Nano, root.Record.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}

	id := path.Base(root.URI)
	page := Page{
		Title:      title,
		Date:       created.Format("02.01.2006"),
		ISODate:    created.Format("2006-01-02"),
		Canonical:  "https://rpne.net/" + id + ".html",
		Paragraphs: body,
	}

	filename := id + ".html"
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	tpl, err := template.New("article").Parse(htmlPage)
	if err != nil {
		log.Fatal(err)
	}

	if err := tpl.Execute(file, page); err != nil {
		file.Close()
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Готово: %s\n", filename)
	fmt.Printf("Частей треда: %d\n", parts)
	fmt.Printf("Заголовок: %s\n", title)
}
