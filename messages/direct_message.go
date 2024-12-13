package messages

import "github.com/asynkron/protoactor-go/actor"

// SendDirectMessage represents a request to send a DM
type SendDirectMessage struct {
    FromUserID string
    ToUserID   string
    Content    string
	ParentID   string
	ActorPID   *actor.PID
}

// SendDirectMessageResponse represents the response to a DM request
type SendDirectMessageResponse struct {
    Success   bool
    MessageID string
    Error     string
    ActorPID  *actor.PID
}

// GetUserMessages represents a request to get all DMs for a user
type GetUserMessages struct {
    UserID string
    ActorPID *actor.PID
}

// GetUserMessagesResponse represents the response containing user's DMs
type GetUserMessagesResponse struct {
    Success   bool
    Messages  []DirectMessage
    Error     string
    ActorPID  *actor.PID
}

// DirectMessage represents a single DM
type DirectMessage struct {
    MessageID  string
    FromUserID string
    ToUserID   string
    Content    string
    Timestamp  int64
	ParentID   string
	ActorPID   *actor.PID
}

