package messages

import "github.com/asynkron/protoactor-go/actor"

type GetFeed struct {
    UserId   string
    ActorPID *actor.PID
}

type FeedResponse struct {
    Success bool
    Error   string
    Feed    []*SubredditFeed
}

type SubredditFeed struct {
    Name        string
    Description string
    Posts       []*PostFeed
}

type PostFeed struct {
    PostId        string
    Title         string
    Content       string
    AuthorId      string
    SubredditName string
    Comments      []*CommentFeed
}

type CommentFeed struct {
    CommentId  string
    Content    string
    AuthorId   string
    Replies    []*CommentFeed
    VoteCount  int
} 