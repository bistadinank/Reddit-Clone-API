package handlers

import (
	"net/http"
	"reddit/messages"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
    enginePID *actor.PID
    system    *actor.ActorSystem
}

func NewCommentHandler(system *actor.ActorSystem, enginePID *actor.PID) *CommentHandler {
    return &CommentHandler{
        enginePID: enginePID,
        system:    system,
    }
}

// Create handles creating a new comment or reply
func (h *CommentHandler) Create(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var request struct {
        PostId   string `json:"postId" binding:"required"`
        ParentId string `json:"parentId"` // Optional, empty for top-level comments
        Content  string `json:"content" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.CreateComment{
        PostId:   request.PostId,
        ParentId: request.ParentId,
        Content:  request.Content,
        AuthorId: username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if createResponse, ok := response.(*messages.CreateCommentResponse); ok {
        if createResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success":   true,
                "commentId": createResponse.CommentId,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   createResponse.Error,
            })
        }
    }
}

// ListByPost handles getting all comments for a post
func (h *CommentHandler) ListByPost(c *gin.Context) {
    postId := c.Param("postId")
    
    msg := &messages.ListPostComments{
        PostId: postId,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if listResponse, ok := response.(*messages.ListPostCommentsResponse); ok {
        if listResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success":  true,
                "comments": listResponse.Comments,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   listResponse.Error,
            })
        }
    }
}

// Vote handles upvoting/downvoting a comment
func (h *CommentHandler) Vote(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    commentId := c.Param("commentId")
    
    var request struct {
        IsUpvote bool `json:"isUpvote"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.Vote{
        UserID:   username.(string),
        TargetID: commentId,
        IsUpvote: request.IsUpvote,
        Type:     "comment",
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if voteResponse, ok := response.(*messages.VoteResponse); ok {
        if voteResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   voteResponse.Error,
            })
        }
    }
}

// Edit handles updating a comment's content
func (h *CommentHandler) Edit(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    commentId := c.Param("commentId")
    
    var request struct {
        Content string `json:"content" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.EditComment{
        CommentId: commentId,
        Content:  request.Content,
        AuthorId: username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if editResponse, ok := response.(*messages.EditCommentResponse); ok {
        if editResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   editResponse.Error,
            })
        }
    }
}

// Delete handles deleting a comment
func (h *CommentHandler) Delete(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    commentId := c.Param("commentId")
    
    msg := &messages.DeleteComment{
        CommentId: commentId,
        AuthorId:  username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if deleteResponse, ok := response.(*messages.DeleteCommentResponse); ok {
        if deleteResponse.Success {
            c.JSON(http.StatusOK, gin.H{"success": true})
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   deleteResponse.Error,
            })
        }
    }
} 