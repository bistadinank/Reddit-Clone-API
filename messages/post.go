package messages

import "github.com/asynkron/protoactor-go/actor"

// Post message for creating a post
type Post struct {
	PostId        string
	Title         string
	Content       string
	AuthorId      string
	SubredditName string
	ActorPID      *actor.PID
}

// PostResponse is the response to a post creation
type PostResponse struct {
	Success bool
	Error   string
	PostId  string
	ActorPID *actor.PID
}

// GetPost message for retrieving a post
type GetPost struct {
	PostId   string
	ActorPID *actor.PID
}

// GetPostResponse is the response to a post retrieval
type GetPostResponse struct {
	Success bool
	Error   string
	Post    *Post
}

// ListSubredditPosts message for getting posts in a subreddit
type ListSubredditPosts struct {
	SubredditName string
	ActorPID      *actor.PID
}

// ListSubredditPostsResponse is the response to a subreddit posts listing
type ListSubredditPostsResponse struct {
	Success bool
	Error   string
	Posts   []*Post
}

// EditPost message for editing a post
type EditPost struct {
	PostId   string
	Title    string
	Content  string
	AuthorId string    // For verification
	ActorPID *actor.PID
}

// EditPostResponse is the response to an edit post request
type EditPostResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}

// DeletePost message for deleting a post
type DeletePost struct {
	PostId   string
	AuthorId string    // For verification
	ActorPID *actor.PID
}

// DeletePostResponse is the response to a delete post request
type DeletePostResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}

// For cascading deletion
type DeletePostComments struct {
	PostId   string
	ActorPID *actor.PID
}

type DeletePostCommentsResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}
