package handlers

import (
	"fmt"
	"net/http"
	"reddit/messages"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gin-gonic/gin"
)

type SubredditHandler struct {
    enginePID *actor.PID
    system    *actor.ActorSystem
}

func NewSubredditHandler(system *actor.ActorSystem, enginePID *actor.PID) *SubredditHandler {
    return &SubredditHandler{
        enginePID: enginePID,
        system:    system,
    }
}

// Create handles subreddit creation
func (h *SubredditHandler) Create(c *gin.Context) {
    // Get username from context
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var request struct {
        Name        string `json:"name" binding:"required"`
        Description string `json:"description" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.CreateSubreddit{
        Name:        request.Name,
        Description: request.Description,
        CreatorId:   username.(string),  // Use the username from token
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if createResponse, ok := response.(*messages.CreateSubredditResponse); ok {
        if createResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "subId":   createResponse.SubId,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   createResponse.Error,
            })
        }
    }
}

// Join handles joining a subreddit
func (h *SubredditHandler) Join(c *gin.Context) {
    subredditName := c.Param("name")
    var request struct {
        UserId string `json:"userId" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.JoinSubreddit{
        SubredditName: subredditName,
        UserId:        request.UserId,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if joinResponse, ok := response.(*messages.JoinSubredditResponse); ok {
        if joinResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "subId":   joinResponse.SubId,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   joinResponse.Error,
            })
        }
    }
}

// Leave handles leaving a subreddit
func (h *SubredditHandler) Leave(c *gin.Context) {
    subredditName := c.Param("name")
    var request struct {
        UserId string `json:"userId" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.LeaveSubreddit{
        SubredditName: subredditName,
        UserId:        request.UserId,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if leaveResponse, ok := response.(*messages.LeaveSubredditResponse); ok {
        if leaveResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   leaveResponse.Error,
            })
        }
    }
}

// List returns available subreddits
func (h *SubredditHandler) List(c *gin.Context) {
    msg := &messages.GetSubreddits{}

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        fmt.Printf("Error getting subreddits: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if listResponse, ok := response.(*messages.GetSubredditsResponse); ok {
        fmt.Printf("Got response: %+v\n", listResponse)
        if listResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success":    true,
                "subreddits": listResponse.Subreddits,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   listResponse.Error,
            })
        }
    }
}

// GetMembers returns members of a subreddit
func (h *SubredditHandler) GetMembers(c *gin.Context) {
    subredditName := c.Param("name")
    
    msg := &messages.GetSubredditMembers{
        SubredditName: subredditName,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if membersResponse, ok := response.(*messages.GetSubredditMembersResponse); ok {
        if membersResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "members": membersResponse.Members,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   membersResponse.Error,
            })
        }
    }
}

// Edit handles updating a subreddit's details
func (h *SubredditHandler) Edit(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    name := c.Param("name")
    
    var request struct {
        Description string `json:"description" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.EditSubreddit{
        Name:        name,
        Description: request.Description,
        AuthorId:    username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if editResponse, ok := response.(*messages.EditSubredditResponse); ok {
        if editResponse.Success {
            c.JSON(http.StatusOK, gin.H{"success": true})
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   editResponse.Error,
            })
        }
    }
} 