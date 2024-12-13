package actors

import (
	"reddit/messages"
	"sort"
	"sync"
	"time"

	"fmt"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/google/uuid"
)

type StoredComment struct {
	CommentId  string
	PostId     string
	ParentId   string
	Content    string
	AuthorId   string
	Timestamp  int64
	Votes     map[string]bool  // username -> isUpvote
}

// Global shared state for all comment actors
var (
	globalComments = make(map[string]*StoredComment)
	postComments = make(map[string][]string)     // PostId -> []CommentId
	commentReplies = make(map[string][]string)   // ParentCommentId -> []CommentId
	commentMutex sync.RWMutex
)

type CommentActor struct {}

func NewCommentActor() *CommentActor {
	return &CommentActor{}
}

func (state *CommentActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.CreateComment:
		response := &messages.CreateCommentResponse{}
		commentId := uuid.New().String()
		
		fmt.Printf("Creating comment: ParentId=%s, PostId=%s\n", msg.ParentId, msg.PostId)
		
		commentMutex.Lock()
		comment := &StoredComment{
			CommentId:  commentId,
			PostId:     msg.PostId,
			ParentId:   msg.ParentId,
			Content:    msg.Content,
			AuthorId:   msg.AuthorId,
			Timestamp:  time.Now().Unix(),
			Votes:      make(map[string]bool),
		}
		globalComments[commentId] = comment
		
		if msg.ParentId == "" {
			fmt.Printf("Adding top-level comment to post %s\n", msg.PostId)
			postComments[msg.PostId] = append(postComments[msg.PostId], commentId)
		} else {
			fmt.Printf("Adding reply to comment %s\n", msg.ParentId)
			commentReplies[msg.ParentId] = append(commentReplies[msg.ParentId], commentId)
		}
		commentMutex.Unlock()
		
		response.Success = true
		response.CommentId = commentId
		response.ActorPID = context.Self()
		context.Respond(response)

	case *messages.ListPostComments:
		fmt.Printf("\nListing comments for post: %s\n", msg.PostId)
		fmt.Printf("Current postComments map: %+v\n", postComments)
		fmt.Printf("Current commentReplies map: %+v\n", commentReplies)
		fmt.Printf("Current globalComments map: %+v\n", globalComments)
		
		response := &messages.ListPostCommentsResponse{}
		response.Comments = state.buildCommentTree(msg.PostId)
		response.Success = true
		context.Respond(response)

	case *messages.Vote:
		response := state.handleVote(msg)
		context.Respond(response)

	case *messages.EditComment:
		response := state.handleEdit(msg)
		context.Respond(response)

	case *messages.DeleteComment:
		response := state.handleDelete(msg)
		context.Respond(response)

	case *messages.DeletePostComments:
		response := &messages.DeletePostCommentsResponse{}
		
		commentMutex.Lock()
		// Get all comments for this post
		if comments, exists := postComments[msg.PostId]; exists {
			// Delete each comment and its replies
			for _, commentId := range comments {
				state.deleteCommentRecursive(commentId)
			}
			delete(postComments, msg.PostId)
			response.Success = true
		} else {
			response.Success = true  // No comments to delete is still a success
		}
		commentMutex.Unlock()
		
		context.Respond(response)
	}
}

// Builds a nested comment tree
func (state *CommentActor) buildCommentTree(postId string) []*messages.Comment {
	commentMutex.RLock()
	defer commentMutex.RUnlock()

	var result []*messages.Comment
	
	// Get top-level comments first
	for _, commentId := range postComments[postId] {
		if comment, exists := globalComments[commentId]; exists {
			commentTree := state.buildCommentWithReplies(comment)
			result = append(result, commentTree)
		}
	}
	
	return result
}

// Recursively builds a comment with its replies
func (state *CommentActor) buildCommentWithReplies(stored *StoredComment) *messages.Comment {
	// Calculate vote count
	voteCount := 0
	for _, isUpvote := range stored.Votes {
		if isUpvote {
			voteCount++
		} else {
			voteCount--
		}
	}

	comment := &messages.Comment{
		CommentId:  stored.CommentId,
		PostId:     stored.PostId,
		ParentId:   stored.ParentId,
		Content:    stored.Content,
		AuthorId:   stored.AuthorId,
		Timestamp:  stored.Timestamp,
		Replies:    make([]*messages.Comment, 0),
		VoteCount:  voteCount,
	}
	
	// Add debug logs
	fmt.Printf("Building comment tree for comment: %s\n", stored.CommentId)
	
	// Check if this comment has any replies
	if replies, exists := commentReplies[stored.CommentId]; exists && len(replies) > 0 {
		fmt.Printf("Found %d replies for comment %s\n", len(replies), stored.CommentId)
		// Add replies recursively
		for _, replyId := range replies {
			if reply, exists := globalComments[replyId]; exists {
				fmt.Printf("Adding reply %s to comment %s\n", replyId, stored.CommentId)
				comment.Replies = append(comment.Replies, state.buildCommentWithReplies(reply))
			}
		}
	} else {
		fmt.Printf("No replies found for comment %s\n", stored.CommentId)
	}
	
	return comment
}

func (state *CommentActor) handleVote(msg *messages.Vote) *messages.VoteResponse {
	commentMutex.Lock()
	defer commentMutex.Unlock()

	fmt.Printf("Handling vote for comment %s by user %s (upvote: %v)\n", 
		msg.TargetID, msg.UserID, msg.IsUpvote)

	if comment, exists := globalComments[msg.TargetID]; exists {
		if comment.Votes == nil {
			comment.Votes = make(map[string]bool)
		}

		// Handle vote change
		if previousVote, hasVoted := comment.Votes[msg.UserID]; hasVoted {
			if previousVote == msg.IsUpvote {
				fmt.Printf("Removing vote from user %s\n", msg.UserID)
				delete(comment.Votes, msg.UserID)
			} else {
				fmt.Printf("Changing vote from user %s\n", msg.UserID)
				comment.Votes[msg.UserID] = msg.IsUpvote
			}
		} else {
			fmt.Printf("Adding new vote from user %s\n", msg.UserID)
			comment.Votes[msg.UserID] = msg.IsUpvote
		}

		fmt.Printf("Current votes for comment %s: %+v\n", msg.TargetID, comment.Votes)
		return &messages.VoteResponse{Success: true}
	}
	return &messages.VoteResponse{Success: false, Error: "Comment not found"}
}

func (state *CommentActor) sortComments(comments []*messages.Comment) {
	sort.Slice(comments, func(i, j int) bool {
		// Sort by vote count first
		iVotes := len(globalComments[comments[i].CommentId].Votes)
		jVotes := len(globalComments[comments[j].CommentId].Votes)
		if iVotes != jVotes {
			return iVotes > jVotes
		}
		// Then by timestamp
		return comments[i].Timestamp > comments[j].Timestamp
	})

	// Sort replies recursively
	for _, comment := range comments {
		state.sortComments(comment.Replies)
	}
}

func (state *CommentActor) handleEdit(msg *messages.EditComment) *messages.EditCommentResponse {
	commentMutex.Lock()
	defer commentMutex.Unlock()

	fmt.Printf("Handling edit for comment %s by user %s\n", msg.CommentId, msg.AuthorId)

	if comment, exists := globalComments[msg.CommentId]; exists {
		// Verify ownership
		if comment.AuthorId != msg.AuthorId {
			return &messages.EditCommentResponse{
				Success: false,
				Error:   "Not authorized to edit this comment",
			}
		}

		// Update content
		comment.Content = msg.Content
		fmt.Printf("Comment %s updated with new content\n", msg.CommentId)
		
		return &messages.EditCommentResponse{Success: true}
	}
	return &messages.EditCommentResponse{Success: false, Error: "Comment not found"}
}

func (state *CommentActor) handleDelete(msg *messages.DeleteComment) *messages.DeleteCommentResponse {
	fmt.Printf("CommentActor: Handling delete for comment %s by user %s\n", msg.CommentId, msg.AuthorId)
	
	commentMutex.Lock()
	defer commentMutex.Unlock()

	comment, exists := globalComments[msg.CommentId]
	if !exists {
		fmt.Printf("CommentActor: Comment %s not found\n", msg.CommentId)
		return &messages.DeleteCommentResponse{
			Success: false,
			Error:   "Comment not found",
		}
	}

	// Verify ownership
	if comment.AuthorId != msg.AuthorId {
		return &messages.DeleteCommentResponse{
			Success: false,
			Error:   "Not authorized to delete this comment",
		}
	}

	// Delete recursively
	state.deleteCommentRecursive(msg.CommentId)

	// Remove from parent's replies if it's a reply
	if comment.ParentId != "" {
		if replies, exists := commentReplies[comment.ParentId]; exists {
			for i, replyId := range replies {
				if replyId == msg.CommentId {
					commentReplies[comment.ParentId] = append(replies[:i], replies[i+1:]...)
					break
				}
			}
		}
	}

	// Remove from post's comments if it's a top-level comment
	if comment.ParentId == "" {
		if comments, exists := postComments[comment.PostId]; exists {
			for i, commentId := range comments {
				if commentId == msg.CommentId {
					postComments[comment.PostId] = append(comments[:i], comments[i+1:]...)
					break
				}
			}
		}
	}

	return &messages.DeleteCommentResponse{Success: true}
}

func (state *CommentActor) deleteCommentRecursive(commentId string) {
	// Delete all replies first
	if replies, exists := commentReplies[commentId]; exists {
		for _, replyId := range replies {
			state.deleteCommentRecursive(replyId)
		}
		delete(commentReplies, commentId)
	}

	// Delete the comment itself
	delete(globalComments, commentId)
}
