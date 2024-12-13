package actors

import (
	"fmt"
	"reddit/messages"
	"sync"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

// Global shared state for all subreddit actors
var (
	globalSubreddits = make(map[string]*Subreddit)
	subredditMutex  sync.RWMutex
)

type SubredditActor struct {
	system *actor.ActorSystem
}

type Subreddit struct {
	Name        string
	Description string
	CreatorId   string
	Members     map[string]bool
}

func NewSubredditActor(system *actor.ActorSystem) *SubredditActor {
	return &SubredditActor{
		system: system,
	}
}

func (state *SubredditActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
		case *messages.CreateSubreddit:
			fmt.Printf("SubredditActor: Creating subreddit %s\n", msg.Name)
			response := &messages.CreateSubredditResponse{}

			subredditMutex.Lock()
			if _, exists := globalSubreddits[msg.Name]; exists {
				subredditMutex.Unlock()
				fmt.Printf("SubredditActor: Subreddit %s already exists\n", msg.Name)
				response.Success = false
				response.Error = "Subreddit already exists"
			} else {
				subreddit := &Subreddit{
					Name:        msg.Name,
					Description: msg.Description,
					CreatorId:   msg.CreatorId,
					Members:     make(map[string]bool),
				}
				globalSubreddits[msg.Name] = subreddit
				subredditMutex.Unlock()
				
				fmt.Printf("SubredditActor: Created subreddit %s. Current subreddits: %v\n", 
					msg.Name, globalSubreddits)
				response.Success = true
				response.SubId = msg.Name
				response.ActorPID = context.Self()
			}

			context.Respond(response)

		case *messages.JoinSubreddit:
			fmt.Printf("SubredditActor: Handling join request for %s by user %s\n", 
				msg.SubredditName, msg.UserId)
			response := &messages.JoinSubredditResponse{}

			subredditMutex.Lock()
			subreddit, exists := globalSubreddits[msg.SubredditName]
			if exists {
				if _, isMember := subreddit.Members[msg.UserId]; isMember {
					fmt.Printf("SubredditActor: User %s is already a member\n", msg.UserId)
					response.Success = false
					response.Error = "User is already a member"
				} else {
					subreddit.Members[msg.UserId] = true
					fmt.Printf("SubredditActor: Added user %s as member. Current members: %v\n", 
						msg.UserId, subreddit.Members)
					response.Success = true
					response.SubId = msg.SubredditName
				}
			} else {
				fmt.Printf("SubredditActor: Subreddit %s not found\n", msg.SubredditName)
				response.Success = false
				response.Error = "Subreddit not found"
			}
			subredditMutex.Unlock()

			context.Respond(response)

		case *messages.GetSubredditMembers:
			fmt.Printf("SubredditActor: Getting members for subreddit %s\n", msg.SubredditName)
			response := &messages.GetSubredditMembersResponse{}

			subredditMutex.RLock()
			subreddit, exists := globalSubreddits[msg.SubredditName]
			fmt.Printf("SubredditActor: Subreddit exists: %v, Members: %v\n", exists, subreddit.Members)
			subredditMutex.RUnlock()

			if exists {
				members := make([]string, 0, len(subreddit.Members))
				for memberId := range subreddit.Members {
					fmt.Printf("SubredditActor: Found member: %s\n", memberId)
					members = append(members, memberId)
				}
				response.Members = members
				response.Success = true
			} else {
					response.Success = false
					response.Error = "Subreddit not found"
			}

			fmt.Printf("SubredditActor: Sending response: %+v\n", response)
			context.Respond(response)

		case *messages.LeaveSubreddit:
			response := &messages.LeaveSubredditResponse{}

			subredditMutex.RLock()
			subreddit, exists := globalSubreddits[msg.SubredditName]
			subredditMutex.RUnlock()

			if exists {
				if _, isMember := subreddit.Members[msg.UserId]; isMember {
					delete(subreddit.Members, msg.UserId)
					response.Success = true
					response.SubId = msg.SubredditName
				} else {
					response.Success = false
					response.Error = "User is not a member of this subreddit"
				}
			} else {
				response.Success = false
				response.Error = "Subreddit not found"
			}

			context.Respond(response)

		case *messages.GetSubreddits:
			fmt.Printf("SubredditActor: Getting all subreddits\n")
			response := &messages.GetSubredditsResponse{}
			response.Success = true

			subredditMutex.RLock()
			allSubreddits := make([]string, 0, len(globalSubreddits))
			for subredditName := range globalSubreddits {
				fmt.Printf("SubredditActor: Found subreddit: %s\n", subredditName)
				allSubreddits = append(allSubreddits, subredditName)
			}
			subredditMutex.RUnlock()

			response.Subreddits = allSubreddits
			context.Respond(response)

		case *messages.DeleteSubreddit:
			response := state.handleDelete(msg)
			context.Respond(response)
	}
}

func (state *SubredditActor) handleDelete(msg *messages.DeleteSubreddit) *messages.DeleteSubredditResponse {
	subredditMutex.Lock()
	defer subredditMutex.Unlock()

	subreddit, exists := globalSubreddits[msg.Name]
	if !exists {
		return &messages.DeleteSubredditResponse{
			Success: false,
			Error:   "Subreddit not found",
		}
	}

	// Verify ownership
	if subreddit.CreatorId != msg.AuthorId {
		return &messages.DeleteSubredditResponse{
			Success: false,
			Error:   "Not authorized to delete this subreddit",
		}
	}

	// Delete all posts in subreddit
	if msg.ActorPID != nil {
		deletePostsMsg := &messages.DeleteSubredditPosts{
			SubredditName: msg.Name,
			ActorPID: msg.ActorPID,
		}
		
		future := state.system.Root.RequestFuture(msg.ActorPID, deletePostsMsg, 5*time.Second)
		if response, err := future.Result(); err != nil {
			return &messages.DeleteSubredditResponse{
				Success: false,
				Error:   "Failed to delete posts: " + err.Error(),
			}
		} else {
			if deleteResponse, ok := response.(*messages.DeleteSubredditPostsResponse); !ok || !deleteResponse.Success {
				return &messages.DeleteSubredditResponse{
					Success: false,
					Error:   "Failed to delete posts: " + deleteResponse.Error,
				}
			}
		}
	}

	// Delete the subreddit itself (members are part of the subreddit struct)
	delete(globalSubreddits, msg.Name)

	return &messages.DeleteSubredditResponse{Success: true}
}

