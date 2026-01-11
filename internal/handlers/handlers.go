package handlers

import (
	"blog_aggregator/external/api"
	"blog_aggregator/internal/cli"
	"blog_aggregator/internal/database"
	"blog_aggregator/internal/state"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func HandlerLogin(s *state.State, cmd cli.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("login requires one argument")
	}

	_, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		return errors.New("user does not exist")
	}

	err = s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Current user set to: %s\n", s.Config.CurrentUserName)

	return nil
}

func HandlerRegister(s *state.State, cmd cli.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("register requires one argument")
	}

	_, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	if err == nil {
		return errors.New("user already exists")
	}

	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	})

	if err != nil {
		return err
	}

	err = s.Config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User created: %s\n", user.Name)

	return nil
}

func HandlerReset(s *state.State, cmd cli.Command) error {
	err := s.Db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("All users deleted")

	return nil

}

func HandlerListUsers(s *state.State, cmd cli.Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	CurrentUserName := s.Config.CurrentUserName

	for _, user := range users {
		if user.Name == CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

func HandlerAggregate(s *state.State, cmd cli.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("agg requires a time duration argument")
	}

	duration, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s\n", duration)

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func HandlerAddFeed(s *state.State, cmd cli.Command, user database.User) error {

	if len(cmd.Args) == 0 || len(cmd.Args) == 1 {
		return errors.New("add feed requires at least two argument")
	}

	nameRss := cmd.Args[0]
	urlRss := strings.ToLower(cmd.Args[1])

	createdFeed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   nameRss,
		Url:    urlRss,
		UserID: user.ID,
	})

	if err != nil {
		return err
	}

	_, err = s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    createdFeed.ID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Created feed: %s\n", createdFeed.Name)

	fmt.Printf("Adding feed: %s\n", nameRss)
	return nil
}

func HandlerListFeeds(s *state.State, cmd cli.Command) error {
	feeds, err := s.Db.GetAllFeedsWithUser(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
	}

	return nil
}

func HandlerFollow(s *state.State, cmd cli.Command, user database.User) error {
	feed, err := s.Db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	createdFeedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		FeedID: feed.ID,
		UserID: user.ID,
	})

	if err != nil {
		return err
	}

	fmt.Printf("Created feed follow: %s and user name: %s\n", createdFeedFollow.FeedName, createdFeedFollow.UserName)

	return nil
}

func HandlerFollowing(s *state.State, cmd cli.Command, user database.User) error {
	feedFollows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	if len(feedFollows) == 0 {
		fmt.Println("No feeds found for this user.")
		return nil
	}

	fmt.Printf("Feeds followed by %s:\n", s.Config.CurrentUserName)
	for _, follow := range feedFollows {
		fmt.Printf("* %s\n", follow.FeedName)
	}

	return nil
}

func HandlerUnfollow(s *state.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) == 0 || len(cmd.Args) == 1 {
		return errors.New("add feed requires at least two argument")
	}

	err := s.Db.UnfollowFeed(context.Background(), database.UnfollowFeedParams{
		UserID: user.ID,
		Url:    cmd.Args[0],
	})
	if err != nil {
		return err
	}

	return nil
}

func scrapeFeeds(s *state.State) error {
	nextFeed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	err = s.Db.MarkFeedFetched(context.Background(), nextFeed.ID)

	if err != nil {
		return err
	}

	fetchedFeed, err := api.FetchRSSFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}

	for i, val := range fetchedFeed.Channel.Item {
		fmt.Printf("%d: %s\n", i, val.Title)
	}

	return nil
}
