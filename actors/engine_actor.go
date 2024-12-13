package actors

import (
	"fmt"
	"reddit/messages"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

type EngineActor struct {
	// Single actors for direct routing
	userPID      *actor.PID
	postPID      *actor.PID
	subredditPID *actor.PID
	commentPID   *actor.PID
	system       *actor.ActorSystem

	// Actor pools for load balancing
	userActors           []*actor.PID
	postActors           []*actor.PID
	subredditActors      []*actor.PID
	directMessageActors  []*actor.PID
	commentActors        []*actor.PID
	currentUserActor     int
	currentCommentActor  int
}

func NewEngineActor(system *actor.ActorSystem) *EngineActor {
	engine := &EngineActor{
		system:              system,
		userActors:         make([]*actor.PID, 0),
		postActors:         make([]*actor.PID, 0),
		subredditActors:    make([]*actor.PID, 0),
		directMessageActors: make([]*actor.PID, 0),
		commentActors:      make([]*actor.PID, 0),
		currentUserActor:   0,
		currentCommentActor: 0,
	}

	// Create main actors
	props := actor.PropsFromProducer(func() actor.Actor { return NewUserActor() })
	engine.userPID = system.Root.Spawn(props)

	props = actor.PropsFromProducer(func() actor.Actor { return NewSubredditActor(system) })
	engine.subredditPID = system.Root.Spawn(props)

	props = actor.PropsFromProducer(func() actor.Actor { return NewPostActor(system) })
	engine.postPID = system.Root.Spawn(props)

	props = actor.PropsFromProducer(func() actor.Actor { return NewCommentActor() })
	engine.commentPID = system.Root.Spawn(props)

	// Create actor pools
	for i := 0; i < 10; i++ {
		// Create subreddit actor
		subredditProps := actor.PropsFromProducer(func() actor.Actor { return NewSubredditActor(system) })
		subredditPID := system.Root.Spawn(subredditProps)
		engine.subredditActors = append(engine.subredditActors, subredditPID)

		// Create user actor
		userProps := actor.PropsFromProducer(func() actor.Actor { return NewUserActor() })
		userPID := system.Root.Spawn(userProps)
		engine.userActors = append(engine.userActors, userPID)

		// Create post actor
		postProps := actor.PropsFromProducer(func() actor.Actor { return NewPostActor(system) })
		postPID := system.Root.Spawn(postProps)
		engine.postActors = append(engine.postActors, postPID)

		// Create direct message actor
		dmProps := actor.PropsFromProducer(func() actor.Actor { return NewDirectMessageActor() })
		dmPID := system.Root.Spawn(dmProps)
		engine.directMessageActors = append(engine.directMessageActors, dmPID)

		// Create comment actor
		commentProps := actor.PropsFromProducer(func() actor.Actor { return NewCommentActor() })
		commentPID := system.Root.Spawn(commentProps)
		engine.commentActors = append(engine.commentActors, commentPID)
	}

	return engine
}

func (state *EngineActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Printf("Engine Actor started\n")

	case *messages.RegisterUser:
		// Forward to next user actor in round-robin fashion
		userActor := state.userActors[state.currentUserActor]
		state.currentUserActor = (state.currentUserActor + 1) % len(state.userActors)
		context.RequestWithCustomSender(userActor, msg, context.Sender())

	case *messages.LoginUser:
		if msg.ActorPID == nil {
			userActor := state.userActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.userActors)
			context.RequestWithCustomSender(userActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.CreateSubreddit:
		fmt.Printf("Engine: Received CreateSubreddit request for %s\n", msg.Name)
		if msg.ActorPID == nil {
			subredditActor := state.subredditActors[state.currentUserActor]
			fmt.Printf("Engine: Routing to subreddit actor %v\n", subredditActor)
			state.currentUserActor = (state.currentUserActor + 1) % len(state.subredditActors)
			context.RequestWithCustomSender(subredditActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.JoinSubreddit:
		if msg.ActorPID == nil {
			subredditActor := state.subredditActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.subredditActors)
			context.RequestWithCustomSender(subredditActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.GetSubreddits:
		fmt.Printf("Engine: Received GetSubreddits request\n")
		if msg.ActorPID == nil {
			subredditActor := state.subredditActors[state.currentUserActor]
			fmt.Printf("Engine: Routing GetSubreddits to actor %v\n", subredditActor)
			state.currentUserActor = (state.currentUserActor + 1) % len(state.subredditActors)
			context.RequestWithCustomSender(subredditActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.LeaveSubreddit:
		if msg.ActorPID == nil {
			subredditActor := state.subredditActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.subredditActors)
			context.RequestWithCustomSender(subredditActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.Post:
		if msg.ActorPID == nil {
			postActor := state.postActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.postActors)
			context.RequestWithCustomSender(postActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.CreateComment:
		fmt.Printf("Engine: Received CreateComment request\n")
		if msg.ActorPID == nil {
			commentActor := state.commentActors[state.currentCommentActor]
			state.currentCommentActor = (state.currentCommentActor + 1) % len(state.commentActors)
			context.RequestWithCustomSender(commentActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.ListPostComments:
		fmt.Printf("Engine: Received ListPostComments request for post %s\n", msg.PostId)
		if msg.ActorPID == nil {
			commentActor := state.commentActors[state.currentCommentActor]
			state.currentCommentActor = (state.currentCommentActor + 1) % len(state.commentActors)
			context.RequestWithCustomSender(commentActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.SendDirectMessage:
		if msg.ActorPID == nil {
			dmActor := state.directMessageActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.directMessageActors)
			context.RequestWithCustomSender(dmActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.Vote:
		if msg.ActorPID == nil {
			switch msg.Type {
			case "post":
				postActor := state.postActors[state.currentUserActor]
				state.currentUserActor = (state.currentUserActor + 1) % len(state.postActors)
				context.RequestWithCustomSender(postActor, msg, context.Sender())
			case "comment":
				commentActor := state.commentActors[state.currentUserActor]
				state.currentUserActor = (state.currentUserActor + 1) % len(state.commentActors)
				context.RequestWithCustomSender(commentActor, msg, context.Sender())
			}
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.GetSubredditMembers:
		fmt.Printf("Engine: Received GetSubredditMembers request for %s\n", msg.SubredditName)
		if msg.ActorPID == nil {
			subredditActor := state.subredditActors[state.currentUserActor]
			fmt.Printf("Engine: Routing GetMembers to actor %v\n", subredditActor)
			state.currentUserActor = (state.currentUserActor + 1) % len(state.subredditActors)
			context.RequestWithCustomSender(subredditActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.ValidateToken:
		fmt.Printf("Engine: Received ValidateToken request for token: %s\n", msg.Token)
		// Forward to user actor for validation
		if msg.ActorPID == nil {
			userActor := state.userActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.userActors)
			context.RequestWithCustomSender(userActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.ListSubredditPosts:
		fmt.Printf("Engine: Received ListSubredditPosts request for %s\n", msg.SubredditName)
		if msg.ActorPID == nil {
			postActor := state.postActors[state.currentUserActor]
			fmt.Printf("Engine: Routing to post actor %v\n", postActor)
			state.currentUserActor = (state.currentUserActor + 1) % len(state.postActors)
			context.RequestWithCustomSender(postActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.GetPost:
		fmt.Printf("Engine: Received GetPost request for post ID: %s\n", msg.PostId)
		if msg.ActorPID == nil {
			postActor := state.postActors[state.currentUserActor]
			fmt.Printf("Engine: Routing to post actor %v\n", postActor)
			state.currentUserActor = (state.currentUserActor + 1) % len(state.postActors)
			context.RequestWithCustomSender(postActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.DeleteComment:
		fmt.Printf("Engine: Received DeleteComment request for comment ID: %s\n", msg.CommentId)
		if msg.ActorPID == nil {
			commentActor := state.commentActors[state.currentCommentActor]
			state.currentCommentActor = (state.currentCommentActor + 1) % len(state.commentActors)
			context.RequestWithCustomSender(commentActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.DeletePost:
		fmt.Printf("Engine: Received DeletePost request for post ID: %s\n", msg.PostId)
		if msg.ActorPID == nil {
			postActor := state.postActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.postActors)
			context.RequestWithCustomSender(postActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.DeleteSubreddit:
		fmt.Printf("Engine: Received DeleteSubreddit request for subreddit: %s\n", msg.Name)
		if msg.ActorPID == nil {
			subredditActor := state.subredditActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.subredditActors)
			context.RequestWithCustomSender(subredditActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.GetFeed:
		fmt.Printf("Engine: Received GetFeed request for user: %s\n", msg.UserId)
		if msg.ActorPID == nil {
			userActor := state.userActors[state.currentUserActor]
			state.currentUserActor = (state.currentUserActor + 1) % len(state.userActors)
			context.RequestWithCustomSender(userActor, msg, context.Sender())
		} else {
			context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
		}

	case *messages.SearchPosts:
		// Forward search request to PostActor
		response, err := state.system.Root.RequestFuture(state.postPID, msg, 5*time.Second).Result()
		if err != nil {
			fmt.Printf("Error searching posts: %v\n", err)
			context.Respond(&messages.SearchPostsResponse{
				Success: false,
				Error:   "Search request timeout",
			})
			return
		}
		context.Respond(response)

	case *messages.EditPost:
		// Forward edit request to PostActor
		response, err := state.system.Root.RequestFuture(state.postPID, msg, 5*time.Second).Result()
		if err != nil {
			fmt.Printf("Error editing post: %v\n", err)
			context.Respond(&messages.EditPostResponse{
				Success: false,
				Error:   "Edit request timeout",
			})
			return
		}
		context.Respond(response)

	case *messages.EditComment:
		// Forward edit request to CommentActor
		response, err := state.system.Root.RequestFuture(state.commentPID, msg, 5*time.Second).Result()
		if err != nil {
			fmt.Printf("Error editing comment: %v\n", err)
			context.Respond(&messages.EditCommentResponse{
				Success: false,
				Error:   "Edit request timeout",
			})
			return
		}
		context.Respond(response)
	}
}
