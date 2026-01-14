package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/strjkc/gator/internal/database"
)

func handlerFetch(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		fmt.Println("Invalid Use: Missing refresh interval!\nExample usage: agg 10s")
		err := fmt.Errorf("Invalid Use: Missing refresh interval!\n")
		return err
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		fmt.Printf("Fetching now!\n")
		rssFeed, dbFeed, err := scrapeFeeds(s)
		if err != nil {
			fmt.Printf("Error fetching feeds: %v\n", err)
			return err
		}
		for _, item := range rssFeed.Channel.Item {
			layout := "Mon, 02 Jan 2006 15:04:05 -0700"
			pubDate, err := time.Parse(layout, item.PubDate)
			if err != nil {
				return err
			}
			params := database.CreatePostParams{ID: uuid.New(), Title: item.Title, Description: sql.NullString{String: item.Description}, Url: item.Link, PublishedAt: sql.NullTime{Time: pubDate}, CreatedAt: time.Now(), UpdatedAt: time.Now(), FeedID: dbFeed.ID}
			_, err = s.db.CreatePost(context.Background(), params)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int
	var err error
	if len(cmd.args) > 0 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
	} else {
		limit = 2
	}

	params := database.GetPostsForUserParams{UserID: user.ID, Limit: limit}
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	for _, post := range posts {
		fmt.Println()
		fmt.Println("####################################")
		fmt.Printf("Title: %v\n", post.Title)
		fmt.Printf("Link: %v\n", post.Url)
		fmt.Printf("Description: %v\n", post.Description)
		fmt.Printf("Published: %v\n", post.PublishedAt)
		fmt.Println("####################################")
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("Invalid Use: Missing Feed url to follow!\nExample usage: follow https://gator.gator.io")
		err := fmt.Errorf("Invalid Use: Missing Feed url to follow!\n")
		return err
	}
	feedUrl := cmd.args[0]
	feedData, err := s.db.GetFeed(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	params := database.CreateFeedFollowParams{CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feedData.ID}
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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		fmt.Println("Invalid Use: Missing Feed Name and URL\nExample usage: addfeed FeedName https://gator.gator.io")
		err := fmt.Errorf("Invalid Use: Missing Feed Name and URL\n")
		return err
	}
	name := cmd.args[0]
	url := cmd.args[1]
	if !strings.Contains(url, "http") {
		fmt.Printf("Invalid Use: Feed URL: \"%s\" is not valid!\nIf the Feed Name contains spaces make sure to wrap the feed name in quotes!\nExample Usage: \"New Feed\" http://somenewfeed.io/rss", url)
		err := fmt.Errorf("Invalid Use: Feed URL is not valid\n")
		return err
	}
	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = handlerFollow(s, command{name: "follow", args: []string{url}}, user)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(feed)
	return nil
}

func scrapeFeeds(s *state) (*RSSFeed, database.Feed, error) {
	data, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return &RSSFeed{}, database.Feed{}, err
	}
	_, err = s.db.MarkedFeedFetched(context.Background(), data.ID)
	if err != nil {
		return &RSSFeed{}, database.Feed{}, err
	}
	feed, err := fetchFeed(context.Background(), data.Url)
	if err != nil {
		return &RSSFeed{}, database.Feed{}, err
	}
	return feed, data, nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("Invalid Use: Missing Feed URL\nExample usage: unfollow https://gator.gator.io")
		err := fmt.Errorf("Invalid Use: Missing Feed URL\n")
		return err
	}
	url := cmd.args[0]
	feedData, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}
	params := database.RemoveFeedFollowParams{UserID: user.ID, FeedID: feedData.ID}
	err = s.db.RemoveFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
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
