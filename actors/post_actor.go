package actors

import (
	"fmt"
	"reddit/messages"
	"sync"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

type StoredPost struct {
	PostId        string
	Title         string
	Content       string
	AuthorId      string
	SubredditName string
	Timestamp     int64
}

// Global shared state for all post actors
var (
	globalPosts = make(map[string]*StoredPost)
	subredditPosts = make(map[string][]string)  // subredditName -> []postId
	postMutex   sync.RWMutex
)

type PostActor struct {
	system *actor.ActorSystem
}

func NewPostActor(system *actor.ActorSystem) *PostActor {
	return &PostActor{
		system: system,
	}
}

func (state *PostActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.Post:
		
		response := &messages.PostResponse{}
		
		// Generate post ID (you might want a better ID generation strategy)
		postId := fmt.Sprintf("post_%s_%s", msg.SubredditName, msg.Title)
		
		postMutex.Lock()
		if _, exists := globalPosts[postId]; exists {
			response.Success = false
			response.Error = "Post already exists"
		} else {
			// Store the post
			post := &StoredPost{
				PostId:        postId,
				Title:         msg.Title,
				Content:       msg.Content,
				AuthorId:      msg.AuthorId,
				SubredditName: msg.SubredditName,
			}
			globalPosts[postId] = post
			
			subredditPosts[msg.SubredditName] = append(subredditPosts[msg.SubredditName], postId)
			
			response.Success = true
			response.PostId = postId
			response.ActorPID = context.Self()
		}
		postMutex.Unlock()
		
		context.Respond(response)

	case *messages.GetPost:
		response := &messages.GetPostResponse{}
		
		postMutex.RLock()
		if post, exists := globalPosts[msg.PostId]; exists {
			response.Success = true
			response.Post = &messages.Post{
				PostId:        post.PostId,
				Title:         post.Title,
				Content:       post.Content,
				AuthorId:      post.AuthorId,
				SubredditName: post.SubredditName,
				ActorPID:      context.Self(),
			}
		} else {
			response.Success = false
			response.Error = "Post not found"
		}
		postMutex.RUnlock()
		
		context.Respond(response)

	case *messages.ListSubredditPosts:
		fmt.Printf("PostActor: Listing posts for subreddit %s\n", msg.SubredditName)
		response := &messages.ListSubredditPostsResponse{}
		response.Posts = make([]*messages.Post, 0)
		
		postMutex.RLock()
		fmt.Printf("PostActor: Found %d total posts\n", len(globalPosts))
		for _, post := range globalPosts {
			if post.SubredditName == msg.SubredditName {
				fmt.Printf("PostActor: Adding post %s to response\n", post.PostId)
				response.Posts = append(response.Posts, &messages.Post{
					PostId:        post.PostId,
					Title:         post.Title,
					Content:       post.Content,
					AuthorId:      post.AuthorId,
					SubredditName: post.SubredditName,
					ActorPID:      context.Self(),
				})
			}
		}
		postMutex.RUnlock()
		
		fmt.Printf("PostActor: Returning %d posts\n", len(response.Posts))
		response.Success = true
		context.Respond(response)

	case *messages.DeletePost:
		response := state.handleDelete(msg)
		context.Respond(response)

	case *messages.DeleteSubredditPosts:
		response := &messages.DeleteSubredditPostsResponse{}
		
		postMutex.Lock()
		// Get all posts for this subreddit
		if posts, exists := subredditPosts[msg.SubredditName]; exists {
			// Delete each post and its comments
			for _, postId := range posts {
				// Delete comments first
				deleteCommentsMsg := &messages.DeletePostComments{
					PostId: postId,
					ActorPID: msg.ActorPID,
				}
				
				future := state.system.Root.RequestFuture(msg.ActorPID, deleteCommentsMsg, 5*time.Second)
				if _, err := future.Result(); err != nil {
					continue // Skip if comment deletion fails
				}
				
				// Delete the post
				delete(globalPosts, postId)
			}
			delete(subredditPosts, msg.SubredditName)
			response.Success = true
		} else {
			response.Success = true // No posts to delete is still a success
		}
		postMutex.Unlock()
		
		context.Respond(response)
	}
}

func (state *PostActor) handleDelete(msg *messages.DeletePost) *messages.DeletePostResponse {
	postMutex.Lock()
	defer postMutex.Unlock()

	post, exists := globalPosts[msg.PostId]
	if !exists {
		return &messages.DeletePostResponse{
			Success: false,
			Error:   "Post not found",
		}
	}

	// Verify ownership
	if post.AuthorId != msg.AuthorId {
		return &messages.DeletePostResponse{
			Success: false,
			Error:   "Not authorized to delete this post",
		}
	}

	// Delete all comments first (send message to comment actor via engine)
	deleteCommentsMsg := &messages.DeletePostComments{
		PostId: msg.PostId,
		ActorPID: msg.ActorPID,
	}
	
	// Send to engine actor and wait for response
	if msg.ActorPID != nil {
		future := state.system.Root.RequestFuture(msg.ActorPID, deleteCommentsMsg, 5*time.Second)
		if response, err := future.Result(); err != nil {
			return &messages.DeletePostResponse{
				Success: false,
				Error:   "Failed to delete comments: " + err.Error(),
			}
		} else {
			if deleteResponse, ok := response.(*messages.DeletePostCommentsResponse); !ok || !deleteResponse.Success {
				return &messages.DeletePostResponse{
					Success: false,
					Error:   "Failed to delete comments: " + deleteResponse.Error,
				}
			}
		}
	}

	// Remove from subreddit's posts
	if posts, exists := subredditPosts[post.SubredditName]; exists {
		for i, postId := range posts {
			if postId == msg.PostId {
				subredditPosts[post.SubredditName] = append(posts[:i], posts[i+1:]...)
				break
			}
		}
	}

	// Delete the post itself
	delete(globalPosts, msg.PostId)

	return &messages.DeletePostResponse{Success: true}
}
