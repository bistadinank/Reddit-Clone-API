package actors

import (
	"github.com/asynkron/protoactor-go/actor"
)

type DirectMessageActor struct {}

func NewDirectMessageActor() *DirectMessageActor {
	return &DirectMessageActor{}
}

func (state *DirectMessageActor) Receive(context actor.Context) {
	// Implement direct message handling
}
