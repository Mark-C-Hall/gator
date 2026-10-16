package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"os"
	"strconv"
	"time"

	"github.com/mark-c-hall/gator/internal/config"
	"github.com/mark-c-hall/gator/internal/database"
	"github.com/mark-c-hall/gator/internal/feed"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type State struct {
	Config *config.Config
	DB     *database.Queries
}

func NewState(cfg *config.Config, db *database.Queries) *State {
	return &State{Config: cfg, DB: db}
}

type Command struct {
	Name string
	Args []string
}

func NewCommand(args []string) Command {
	cmd := Command{Name: args[0]}
	if len(args) > 1 {
		cmd.Args = args[1:]
	}
	return cmd
}

type Commands struct {
	m map[string]func(*State, Command) error
}

func NewCommands() *Commands {
	return &Commands{m: make(map[string]func(*State, Command) error)}
}

func (c *Commands) Run(s *State, cmd Command) error {
	if handler, ok := c.m[cmd.Name]; ok {
		return handler(s, cmd)
	}
	return errors.New("command not found in registered command")
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.m[name] = f
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("login handler expects a single argument: username")
	}
	username := cmd.Args[0]

	_, err := s.DB.GetUser(context.Background(), username)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user is not registered in database")
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("login handler: error setting username: %w", err)
	}
	fmt.Printf("User '%s' has been set\n", username)
	return nil
}

func HandlderRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("register handler expects a single argument: username")
	}
	username := cmd.Args[0]

	_, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		UserID:    uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == pqerror.UniqueViolation {
		return errors.New("username already exists")
	}
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("error setting username: %w", err)
	}

	fmt.Printf("User '%s' was created\n", username)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.DB.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}
	fmt.Println("All users deleted")
	return nil
}

func HandlerGetUsers(s *State, cmd Command) error {
	users, err := s.DB.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %w", err)
	}
	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func scrapeFeeds(s *State) error {
	nextFeed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error getting next feed to fetch: %w", err)
	}
	feed, err := feed.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}
	err = s.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		FeedID:        nextFeed.FeedID,
		LastFetchedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("error marking feed fetched: %w", err)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		err = addPost(s, item, nextFeed.FeedID)
		if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == pqerror.UniqueViolation {
			continue
		} else {
			return fmt.Errorf("error adding post: %w", err)
		}
	}
	return nil
}

func parsePubDate(pubDate string) (time.Time, error) {
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, pubDate); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized pubDate format: %q", pubDate)
}

func addPost(s *State, post feed.RSSItem, feedID uuid.UUID) error {
	pubDate, err := parsePubDate(post.PubDate)
	if err != nil {
		return fmt.Errorf("error parsing post pubDate: %w", err)
	}
	_, err = s.DB.CreatePost(context.Background(), database.CreatePostParams{
		PostID:      uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       post.Title,
		Url:         post.Link,
		Description: post.Description,
		PublishedAt: pubDate,
		FeedID:      feedID,
	})
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == pqerror.UniqueViolation {
		fmt.Printf("URL %s already exists, skipping...\n", post.Link)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating post: %w", err)
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return errors.New("agg expects one argument: timeBetweenRequests")
	}
	time_between_reqs := cmd.Args[0]
	duration, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("error parsing time between requests string: %w", err)
	}
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("addfeed expects two arguments: name and url")
	}
	name, url := cmd.Args[0], cmd.Args[1]
	newFeed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		FeedID:    uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.UserID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}
	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		FeedFollowID: uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UserID:       user.UserID,
		FeedID:       newFeed.FeedID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}
	fmt.Println(newFeed)
	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	feeds, err := s.DB.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all feeds from db: %w", err)
	}
	for _, feed := range feeds {
		user, err := s.DB.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user ID, continuing\n")
			continue
		}
		fmt.Printf("Feed name: %s, Feed URL: %s, Username: %s\n", feed.Name, feed.Url, user.Name)
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("follow expects one argument: url")
	}
	url := cmd.Args[0]
	feed, err := s.DB.GetFeedByURL(context.Background(), url)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking if feed exists: %w", err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("feed does exist with the given url")
	}
	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		FeedFollowID: uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UserID:       user.UserID,
		FeedID:       feed.FeedID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}
	fmt.Printf("feed: %s, name: %s\n", feed.Name, user.Name)
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	feed_follows, err := s.DB.GetFeedFollowsForUser(context.Background(), user.UserID)
	if err != nil {
		return fmt.Errorf("error getting feed follows: %w", err)
	}
	for _, ff := range feed_follows {
		fmt.Printf("Feed name: %s\n", ff.FeedName)
	}
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("follow expects one argument: url")
	}
	url := cmd.Args[0]
	feed, err := s.DB.GetFeedByURL(context.Background(), url)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking if feed exists: %w", err)
	}
	err = s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.UserID,
		FeedID: feed.FeedID,
	})
	if err != nil {
		return fmt.Errorf("error deleting feed follow: %w", err)
	}
	return nil
}

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Args) > 0 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("error parsing limit argument: %w", err)
		}
		limit = n
	}

	posts, err := s.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.UserID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting user's posts: %w", err)
	}
	for _, content := range posts {
		fmt.Printf("Title: %s\n", content.Title)
		fmt.Printf("Content: %s\n", content.Description)
	}
	return nil
}
