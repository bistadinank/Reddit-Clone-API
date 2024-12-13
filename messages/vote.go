package messages

import "github.com/asynkron/protoactor-go/actor"

type Vote struct {
    UserID    string
    TargetID  string
    IsUpvote  bool
    Type      string    // "post" or "comment"
    ActorPID  *actor.PID
}

type VoteResponse struct {
    Success bool
    Error   string
	ActorPID *actor.PID
}
