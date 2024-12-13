package main

import (
	"reddit/actors"
	"reddit/api/handlers"
	"reddit/api/routes"

	"github.com/asynkron/protoactor-go/actor"
)

func main() {
	// Initialize actor system
	system := actor.NewActorSystem()
	
	// Create engine actor with the system
	engineProps := actor.PropsFromProducer(func() actor.Actor { return actors.NewEngineActor(system) })
	enginePID := system.Root.Spawn(engineProps)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(system, enginePID)
	subredditHandler := handlers.NewSubredditHandler(system, enginePID)
	postHandler := handlers.NewPostHandler(system, enginePID)
	commentHandler := handlers.NewCommentHandler(system, enginePID)

	// Setup router with system and enginePID
	router := routes.SetupRouter(system, enginePID, userHandler, subredditHandler, postHandler, commentHandler)

	// Run HTTP server
	router.Run(":8080")
}
