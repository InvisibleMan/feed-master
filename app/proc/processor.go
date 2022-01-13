// Package proc provided the primary blocking loop
// updating from sources and making feeds
package proc

import (
	"context"
	"regexp"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/syncs"

	"github.com/umputun/feed-master/app/feed"
	"github.com/umputun/feed-master/app/models"
)

// TelegramNotif is interface to send messages to telegram
type TelegramNotif interface {
	Send(chanID string, item feed.Item) error
}

// Processor is a feed reader and store writer
type Processor struct {
	Conf        *Conf
	Store       *BoltDB
	TelegramBot *TelegramClientV2
}

// Conf for feeds config yml
type Conf struct {
	Feeds  map[string]Feed `yaml:"feeds"`
	System struct {
		UpdateInterval time.Duration `yaml:"update"`
		MaxItems       int           `yaml:"max_per_feed"`
		MaxTotal       int           `yaml:"max_total"`
		MaxKeepInDB    int           `yaml:"max_keep"`
		Concurrent     int           `yaml:"concurrent"`
		BaseURL        string        `yaml:"base_url"`
	} `yaml:"system"`
}

// Feed defines config section for a feed~
type Feed struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Link        string `yaml:"link"`
	Image       string `yaml:"image"`
	Language    string `yaml:"language"`
	Filter      Filter `yaml:"filter"`
	Sources     []struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	} `yaml:"sources"`
	ExtendDateTitle string `yaml:"ext_date"`
}

// Filter defines feed section for a feed filter~
type Filter struct {
	Title string `yaml:"title"`
}

type FeedsStore interface {
	Iterate(func(feed models.Feed) error) error
}

// Do activates loop of goroutine for each feed, concurrency limited by p.Conf.Concurrent
func (p *Processor) Do(feedIterator FeedsStore) {
	log.Printf("[INFO] activate processor")
	p.setDefaults()

	for {
		swg := syncs.NewSizedGroup(p.Conf.System.Concurrent, syncs.Preemptive)
		feedIterator.Iterate(func(f models.Feed) error {
			swg.Go(func(context.Context) {
				log.Printf("[INFO] fetch feed: '%s'", f.URL)

				// p.feed(f, p.Conf.System.MaxItems)
			})

			return nil
		})

		// for name, fm := range p.Conf.Feeds {
		// 	// keep up to MaxKeepInDB items in bucket
		// 	if removed, err := p.Store.removeOld(name, p.Conf.System.MaxKeepInDB); err == nil {
		// 		if removed > 0 {
		// 			log.Printf("[DEBUG] removed %d from %s", removed, name)
		// 		}
		// 	} else {
		// 		log.Printf("[WARN] failed to remove, %v", err)
		// 	}
		// }

		swg.Wait()
		log.Printf("[DEBUG] refresh completed. Next iteration after: '%v'", p.Conf.System.UpdateInterval)

		time.Sleep(p.Conf.System.UpdateInterval)
	}
}

func (p *Processor) feed(f Feed, max int) {
	rss, err := feed.Parse(f.Link)
	if err != nil {
		log.Printf("[WARN] failed to parse %s, %v", f.Link, err)
		return
	}

	// up to MaxItems (5) items from each feed
	upto := max
	if len(rss.ItemList) <= max {
		upto = len(rss.ItemList)
	}

	for _, item := range rss.ItemList[:upto] {
		// skip 1y and older
		if item.DT.Before(time.Now().AddDate(-1, 0, 0)) {
			continue
		}

		name := f.Title
		created, err := p.Store.Save(name, item)
		if err != nil {
			log.Printf("[WARN] failed to save %s (%s) to %s, %v", item.GUID, item.PubDate, name, err)
		}

		if !created {
			return
		}

		// if err := p.TelegramNotif.Send("telegramChannel", item); err != nil {
		// 	log.Printf("[WARN] failed to send telegram message, url=%s to channel=%s, %v",
		// 		item.Enclosure.URL, "telegramChannel", err)
		// }
	}
}

func (p *Processor) setDefaults() {
	if p.Conf.System.Concurrent == 0 {
		p.Conf.System.Concurrent = 8
	}
	if p.Conf.System.MaxItems == 0 {
		p.Conf.System.MaxItems = 5
	}
	if p.Conf.System.MaxTotal == 0 {
		p.Conf.System.MaxTotal = 100
	}
	if p.Conf.System.MaxKeepInDB == 0 {
		p.Conf.System.MaxKeepInDB = 5000
	}
	if p.Conf.System.UpdateInterval == 0 {
		p.Conf.System.UpdateInterval = time.Minute * 5
	}
}

func (filter *Filter) skip(item feed.Item) (bool, error) {
	if filter.Title != "" {
		matched, err := regexp.MatchString(filter.Title, item.Title)
		if err != nil {
			return matched, err
		}
		if matched {
			return true, err
		}
	}

	return false, nil
}
