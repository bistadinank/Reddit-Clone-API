package handlers

import (
	"net/http"
	"reddit/messages"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gin-gonic/gin"
)

type PostHandler struct {
    enginePID *actor.PID
    system    *actor.ActorSystem
}

func NewPostHandler(system *actor.ActorSystem, enginePID *actor.PID) *PostHandler {
    return &PostHandler{
        enginePID: enginePID,
        system:    system,
    }
}

// Create handles post creation
func (h *PostHandler) Create(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var request struct {
        Title       string `json:"title" binding:"required"`
        Content     string `json:"content" binding:"required"`
        SubredditName string `json:"subredditName" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.Post{
        Title:       request.Title,
        Content:     request.Content,
        AuthorId:    username.(string),
        SubredditName: request.SubredditName,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if postResponse, ok := response.(*messages.PostResponse); ok {
        if postResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "postId":  postResponse.PostId,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   postResponse.Error,
            })
        }
    }
}

// Get handles retrieving a single post
func (h *PostHandler) Get(c *gin.Context) {
    postId := c.Param("postId")
    
    msg := &messages.GetPost{
        PostId: postId,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if getResponse, ok := response.(*messages.GetPostResponse); ok {
        if getResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "post":    getResponse.Post,
            })
        } else {
            c.JSON(http.StatusNotFound, gin.H{
                "success": false,
                "error":   getResponse.Error,
            })
        }
    }
}

// ListBySubreddit handles listing posts in a subreddit
func (h *PostHandler) ListBySubreddit(c *gin.Context) {
    subredditName := c.Param("name")
    
    msg := &messages.ListSubredditPosts{
        SubredditName: subredditName,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if listResponse, ok := response.(*messages.ListSubredditPostsResponse); ok {
        if listResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "posts":   listResponse.Posts,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   listResponse.Error,
            })
        }
    }
}

// Edit handles updating a post's content
func (h *PostHandler) Edit(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    postId := c.Param("postId")
    
    var request struct {
        Title   string `json:"title,omitempty"`
        Content string `json:"content,omitempty"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.EditPost{
        PostId:   postId,
        Title:    request.Title,
        Content:  request.Content,
        AuthorId: username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if editResponse, ok := response.(*messages.EditPostResponse); ok {
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

// Delete handles deleting a post
func (h *PostHandler) Delete(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    postId := c.Param("postId")
    
    msg := &messages.DeletePost{
        PostId:   postId,
        AuthorId: username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if deleteResponse, ok := response.(*messages.DeletePostResponse); ok {
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

// Search handles searching for posts
func (h *PostHandler) Search(c *gin.Context) {
    query := c.Query("q")  // Get search query from URL parameter
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
        return
    }

    msg := &messages.SearchPosts{
        Query: query,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if searchResponse, ok := response.(*messages.SearchPostsResponse); ok {
        if searchResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "posts":   searchResponse.Posts,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   searchResponse.Error,
            })
        }
    }
} 