package messages

import "time"

type MetricsMessage struct {
	Action       string
	Success      bool
	ResponseTime time.Duration
	Error        string
}

type ActionType string

const (
	ActionRegister       ActionType = "register"
	ActionLogin         ActionType = "login"
	ActionCreatePost    ActionType = "create_post"
	ActionCreateComment ActionType = "create_comment"
	ActionVote         ActionType = "vote"
	ActionJoinSubreddit ActionType = "join_subreddit"
	ActionCreateSubreddit ActionType = "create_subreddit"
	ActionLeaveSubreddit  ActionType = "leave_subreddit"
	ActionSendDM        ActionType = "send_dm"
)
