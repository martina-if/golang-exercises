package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type SeenUrls struct {
	urls map[string]bool
	mux  sync.Mutex
}

var seenUrls SeenUrls

func (u *SeenUrls) seen(url string) {
	u.mux.Lock()
	defer u.mux.Unlock()
	u.urls[url] = true
}

func (u *SeenUrls) isSeen(url string) bool {
	u.mux.Lock()
	defer u.mux.Unlock()
	return u.urls[url]
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func doCrawl(url string, depth int, fetcher Fetcher, ch chan string) {
	defer close(ch)
	if depth <= 0 {
		return
	}
	_, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	if seenUrls.isSeen(url) {
		return
	}
	seenUrls.seen(url)
	//fmt.Printf("found: %s %q\n", url, body)
	ch <- url

	channels := make([]chan string, len(urls))
	for i, u := range urls {
		channels[i] = make(chan string)
		go doCrawl(u, depth-1, fetcher, channels[i])
	}

	for i := range channels {
		for u := range channels[i] {
			ch <- u
		}
	}
	return
}

func Crawl(url string, depth int, fetcher Fetcher, ch chan string) {
	doCrawl(url, depth, fetcher, ch)
}

func main() {
	seenUrls = SeenUrls{
		urls: make(map[string]bool),
	}
	ch := make(chan string)
	go Crawl("http://golang.org/", 4, fetcher, ch)
	for url := range ch {
		fmt.Println(url)
	}

}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

