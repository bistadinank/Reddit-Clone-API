package actors

import (
	"fmt"
	"reddit/messages"
	"strings"
	"sync"

	"github.com/asynkron/protoactor-go/actor"
)

// Global shared state for all user actors
var (
	globalUsers = make(map[string]string) // username -> password
	globalKarma = make(map[string]int)    // userID -> karma
	userMutex   sync.RWMutex
)

type UserActor struct {}

func NewUserActor() *UserActor {
	return &UserActor{}
}

func (state *UserActor) ValidateToken(token string) (string, bool) {
	// Token format: "reddit-token-username"
	parts := strings.Split(token, "-")
	if len(parts) != 3 || parts[0] != "reddit" || parts[1] != "token" {
		return "", false
	}

	username := parts[2]
	
	// Check if user exists
	userMutex.RLock()
	_, exists := globalUsers[username]
	userMutex.RUnlock()

	return username, exists
}

func (state *UserActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
		case *messages.RegisterUser:
			fmt.Printf("UserActor: Registration attempt for user: %s\n", msg.Username)
			response := &messages.RegisterUserResponse{}

			userMutex.Lock()
			if _, exists := globalUsers[msg.Username]; exists {
				userMutex.Unlock()
				fmt.Printf("UserActor: Username %s already exists\n", msg.Username)
				response.Success = false
				response.Error = "Username already exists"
			} else {
				globalUsers[msg.Username] = msg.Password
				userMutex.Unlock()
				fmt.Printf("UserActor: Successfully registered user %s. Current users: %v\n", 
					msg.Username, globalUsers)
				response.Success = true
				response.UserId = msg.Username
				response.ActorPID = context.Self()
			}

			context.Respond(response)

		case *messages.LoginUser:
			fmt.Printf("UserActor: Login attempt for user: %s\n", msg.Username)
			response := &messages.LoginUserResponse{}

			userMutex.RLock()
			storedPassword, exists := globalUsers[msg.Username]
			userMutex.RUnlock()

			if exists {
				fmt.Printf("UserActor: User exists, checking password\n")
				if storedPassword == msg.Password {
					response.Success = true
					response.Token = "reddit-token-" + msg.Username
					fmt.Printf("UserActor: Login successful\n")
				} else {
					response.Success = false
					response.Error = "Invalid credentials"
					fmt.Printf("UserActor: Password mismatch\n")
				}
			} else {
				response.Success = false
				response.Error = "Invalid credentials"
				fmt.Printf("UserActor: User not found\n")
			}

			context.Respond(response)

		case *messages.UpdateKarma:
			userMutex.Lock()
			if _, exists := globalUsers[msg.UserID]; exists {
				globalKarma[msg.UserID] += msg.Change
			}
			userMutex.Unlock()

		case *messages.GetKarma:
			userMutex.RLock()
			if _, exists := globalUsers[msg.UserID]; !exists {
				userMutex.RUnlock()
				context.Respond(&messages.GetKarmaResponse{
					Success: false,
					Error:   "User not found",
				})
				return
			}

			context.Respond(&messages.GetKarmaResponse{
				Success: true,
				Karma:   globalKarma[msg.UserID],
			})
			userMutex.RUnlock()

		case *messages.ValidateToken:
			username, valid := state.ValidateToken(msg.Token)
			response := &messages.ValidateTokenResponse{
				Success:  valid,
				Username: username,
			}
			if !valid {
				response.Error = "Invalid token"
			}
			context.Respond(response)
	}
}
