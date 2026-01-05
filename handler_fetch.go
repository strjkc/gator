package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/strjkc/gator/internal/database"
)

func handlerFetch(s *state, cmd command) error {
	data, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(data.Channel.Title)
	fmt.Println(data.Channel.Link)
	fmt.Println(data.Channel.Description)
	for _, item := range data.Channel.Item {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.Description)
		fmt.Println(item.PubDate)
	}
	return nil
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

func handlerGetFeeds(s *state, cmd command) error {
	data, err := s.db.GetFeedsAndUsers(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, item := range data {
		fmt.Printf("Feed Name: %s\n"+
			"URL: %s\n"+
			"Username: %s\n",
			item.Name, item.Url, item.Name_2)
	}
	return nil
}

func handlerFollow(s *state, cmd command) error {
	feedUrl := cmd.args[0]
	feedData, err := s.db.GetFeed(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	userData, err := s.db.GetUser(context.Background(), s.conf.User)
	if err != nil {
		return err
	}
	params := database.CreateFeedFollowParams{CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: userData.ID, FeedID: feedData.ID}
	data, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("User: %s Feed: %s\n", data[0].Name_2, data[0].Name)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	data, err := s.db.GetFeedFollowsForUser(context.Background(), s.conf.User)
	if err != nil {
		return err
	}
	for _, item := range data {
		fmt.Println(item.Name)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	name := cmd.args[0]
	url := cmd.args[1]
	currUser := s.conf.User
	userData, err := s.db.GetUser(context.Background(), currUser)
	if err != nil {
		return err
	}
	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    userData.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = handlerFollow(s, command{name: "follow", args: []string{url}})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(feed)
	return nil
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rssFeed := &RSSFeed{}
	if err := xml.NewDecoder(resp.Body).Decode(rssFeed); err != nil {
		return nil, err
	}
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i := range rssFeed.Channel.Item {
		rssItem := &rssFeed.Channel.Item[i]
		rssItem.Title = html.UnescapeString(rssItem.Title)
		rssItem.Description = html.UnescapeString(rssItem.Description)
	}
	return rssFeed, nil
}
