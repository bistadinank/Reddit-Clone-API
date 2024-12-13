package messages

import "github.com/asynkron/protoactor-go/actor"

// User related messages
type RegisterUser struct {
	Username string
	Password string
}

type RegisterUserResponse struct {
	Success bool
	Error   string
	UserId  string
	ActorPID *actor.PID
}

type LoginUser struct {
	Username string
	Password string
	ActorPID *actor.PID
}

type LoginUserResponse struct {
	Success bool
	Error   string
	Token   string
}

// ValidateToken message to validate a token
type ValidateToken struct {
	Token    string
	ActorPID *actor.PID
}

// ValidateTokenResponse is the response to a token validation request
type ValidateTokenResponse struct {
	Success  bool
	Username string
	Error    string
}

type EditUserProfile struct {
	UserId      string
	Email       string
	DisplayName string
	Bio         string
	ActorPID    *actor.PID
}

type EditUserProfileResponse struct {
	Success  bool
	Error    string
	ActorPID *actor.PID
}
