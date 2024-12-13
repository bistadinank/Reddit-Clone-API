package handlers

import (
	"net/http"
	"reddit/messages"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
    enginePID *actor.PID
    system    *actor.ActorSystem
}

func NewUserHandler(system *actor.ActorSystem, enginePID *actor.PID) *UserHandler {
    return &UserHandler{
        enginePID: enginePID,
        system:    system,
    }
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
    var request struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.RegisterUser{
        Username: request.Username,
        Password: request.Password,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if registerResponse, ok := response.(*messages.RegisterUserResponse); ok {
        if registerResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "userId": registerResponse.UserId,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   registerResponse.Error,
            })
        }
    }
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
    var request struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.LoginUser{
        Username: request.Username,
        Password: request.Password,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if loginResponse, ok := response.(*messages.LoginUserResponse); ok {
        if loginResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "token":   loginResponse.Token,
            })
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   loginResponse.Error,
            })
        }
    }
}

// GetKarma handles karma retrieval
func (h *UserHandler) GetKarma(c *gin.Context) {
    userId := c.Param("userId")
    
    msg := &messages.GetKarma{
        UserID: userId,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if karmaResponse, ok := response.(*messages.GetKarmaResponse); ok {
        if karmaResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "karma":   karmaResponse.Karma,
            })
        } else {
            c.JSON(http.StatusNotFound, gin.H{
                "success": false,
                "error":   karmaResponse.Error,
            })
        }
    }
}

// EditProfile handles updating a user's profile
func (h *UserHandler) EditProfile(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    
    var request struct {
        Email       string `json:"email,omitempty"`
        DisplayName string `json:"displayName,omitempty"`
        Bio         string `json:"bio,omitempty"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    msg := &messages.EditUserProfile{
        UserId:      username.(string),
        Email:       request.Email,
        DisplayName: request.DisplayName,
        Bio:        request.Bio,
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if editResponse, ok := response.(*messages.EditUserProfileResponse); ok {
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

// GetFeed handles user feed retrieval
func (h *UserHandler) GetFeed(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    msg := &messages.GetFeed{
        UserId: username.(string),
    }

    response, err := h.system.Root.RequestFuture(h.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Request timeout"})
        return
    }

    if feedResponse, ok := response.(*messages.FeedResponse); ok {
        if feedResponse.Success {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "feed":    feedResponse.Feed,
            })
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   feedResponse.Error,
            })
        }
    }
} 