package messages

import "github.com/asynkron/protoactor-go/actor"

// Subreddit related messages
type CreateSubreddit struct {
	Name        string
	Description string
	CreatorId   string
	ActorPID    *actor.PID
}

type CreateSubredditResponse struct {
	Success bool
	Error   string
	SubId     string
	ActorPID  *actor.PID
}



type JoinSubreddit struct {
	UserId    string
	SubredditName string
	ActorPID      *actor.PID
}

type JoinSubredditResponse struct {
	Success bool
	Error   string
	SubId   string
}

type GetSubredditMembers struct {
	SubredditName string
	ActorPID      *actor.PID
}

type GetSubredditMembersResponse struct {
	Members []string
	Success bool
	Error   string
}

type LeaveSubreddit struct {
	UserId        string
	SubredditName string
	ActorPID      *actor.PID
}

type LeaveSubredditResponse struct {
	Success bool
	Error   string
	SubId   string
}

type GetSubredditPosts struct {
	SubredditName string
	ActorPID      *actor.PID
}

type GetSubredditPostsResponse struct {
	Success bool
	Error   string
	Posts   []Post
}

type GetSubreddits struct {
	ActorPID *actor.PID
}

type GetSubredditsResponse struct {
	Success    bool
	Subreddits []string
	Error      string
}

type EditSubreddit struct {
	Name        string
	Description string
	AuthorId    string    // Only moderators/creators can edit
	ActorPID    *actor.PID
}

type EditSubredditResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}

type DeleteSubreddit struct {
	Name     string
	AuthorId string    // Only creator can delete
	ActorPID *actor.PID
}

type DeleteSubredditResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}

// For cascading deletion
type DeleteSubredditPosts struct {
	SubredditName string
	ActorPID      *actor.PID
}

type DeleteSubredditPostsResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}
