package messages

import "github.com/asynkron/protoactor-go/actor"

type UpdateKarma struct {
    UserID string
    Change int
}

type GetKarma struct {
    UserID string
    ActorPID *actor.PID
}

type GetKarmaResponse struct {
    Success bool
    Karma   int
    Error   string
    ActorPID *actor.PID
}
