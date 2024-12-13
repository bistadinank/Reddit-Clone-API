package routes

import (
	"reddit/api/handlers"
	"reddit/api/middleware"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gin-gonic/gin"
)

func SetupRouter(system *actor.ActorSystem, enginePID *actor.PID, 
                userHandler *handlers.UserHandler, 
                subredditHandler *handlers.SubredditHandler,
                postHandler *handlers.PostHandler,
                commentHandler *handlers.CommentHandler) *gin.Engine {
    router := gin.Default()
    
    // Public routes
    router.POST("/register", userHandler.Register)
    router.POST("/login", userHandler.Login)

    // Protected routes
    authorized := router.Group("/")
    authorized.Use(middleware.NewAuthMiddleware(system, enginePID))
    {
        authorized.GET("/user/:userId/karma", userHandler.GetKarma)
        authorized.POST("/subreddit", subredditHandler.Create)
        authorized.POST("/subreddit/:name/join", subredditHandler.Join)
        authorized.POST("/subreddit/:name/leave", subredditHandler.Leave)
        authorized.GET("/subreddits", subredditHandler.List)
        authorized.GET("/subreddit/:name/members", subredditHandler.GetMembers)
        authorized.POST("/post", postHandler.Create)
        authorized.GET("/post/:postId", postHandler.Get)
        authorized.GET("/subreddit/:name/posts", postHandler.ListBySubreddit)
        authorized.POST("/comment", commentHandler.Create)
        authorized.GET("/post/:postId/comments", commentHandler.ListByPost)
        authorized.POST("/comment/:commentId/vote", commentHandler.Vote)
        authorized.PATCH("/comment/:commentId", commentHandler.Edit)
        authorized.PATCH("/post/:postId", postHandler.Edit)
        authorized.PATCH("/subreddit/:name", subredditHandler.Edit)
        authorized.PATCH("/user/profile", userHandler.EditProfile)
        authorized.DELETE("/comment/:commentId", commentHandler.Delete)
        authorized.DELETE("/post/:postId", postHandler.Delete)
        
    }

    return router
} 