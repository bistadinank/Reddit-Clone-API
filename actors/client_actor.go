package actors

import (
	"fmt"
	"math/rand"
	"reddit/messages"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

type ClientActor struct {
    enginePID     *actor.PID
		controllerPID *actor.PID
    rand          *rand.Rand
    username      string
    token         string
    mySubreddits  []string    // subreddits created/joined by this user
    myPosts       []string    // posts created by this user
    myComments    []string    // comments created by this user
		myDms         []string    // direct messages sent by this user
    actionDelay   time.Duration
		startTime     time.Time
		userToActorPID map[string]*actor.PID
}

// Add this helper function at the bottom of the file
func contains(slice []string, item string) bool {
	for _, s := range slice {
			if s == item {
					return true
			}
	}
	return false
}

func NewClientActor(enginePID *actor.PID, controllerPID *actor.PID) *ClientActor {
    return &ClientActor{
        enginePID:    enginePID,
				controllerPID: controllerPID,
        rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
        mySubreddits: make([]string, 0),
        myPosts:      make([]string, 0),
        myComments:   make([]string, 0),
				myDms:        make([]string, 0),
        actionDelay:  time.Duration(rand.Intn(1000)) * time.Millisecond,
				startTime:    time.Now(),
				userToActorPID: make(map[string]*actor.PID),
    }
}

func (state *ClientActor) generateContent() string {
    return fmt.Sprintf("content_%d", state.rand.Intn(10000))
}

func (state *ClientActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *actor.Started:
        // Register user when actor starts
        state.register(context)

    case *messages.RegisterUserResponse:
			metricsMsg := &messages.MetricsMessage{
				Action:       "register",
				Success:      msg.Success,
				ResponseTime: time.Since(state.startTime),
				Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)
			if msg.Success {
					state.username = msg.UserId
					// Automatically login after successful registration
					state.userToActorPID["user"] = msg.ActorPID
					state.login(context)
			}

    case *messages.LoginUserResponse:
			metricsMsg := &messages.MetricsMessage{
				Action:       "login",
				Success:      msg.Success,
				ResponseTime: time.Since(state.startTime),
				Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)
        if msg.Success {
            state.token = msg.Token
            // Start periodic actions after successful login
            context.SetReceiveTimeout(state.actionDelay)
        }

    case *actor.ReceiveTimeout:
        if state.token != "" {
            state.performRandomAction(context)
        }
        context.SetReceiveTimeout(state.actionDelay)

    case *messages.CreateSubredditResponse:
			metricsMsg := &messages.MetricsMessage{
				Action:       "create_subreddit",
				Success:      msg.Success,
				ResponseTime: time.Since(state.startTime),
				Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)

			if msg.Success {
					state.userToActorPID["subreddit"] = msg.ActorPID
					state.mySubreddits = append(state.mySubreddits, msg.SubId)
				}

		case *messages.JoinSubredditResponse:
			metricsMsg := &messages.MetricsMessage{
					Action:       "join_subreddit",
					Success:      msg.Success,
					ResponseTime: time.Since(state.startTime),
					Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)

			if msg.Success {
				state.mySubreddits = append(state.mySubreddits, msg.SubId)
			}

		case *messages.LeaveSubredditResponse:
			metricsMsg := &messages.MetricsMessage{
					Action:       "leave_subreddit",
					Success:      msg.Success,
					ResponseTime: time.Since(state.startTime),
					Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)

			if msg.Success {
				// Remove the subreddit from the user's list
				for i, sub := range state.mySubreddits {
					if sub == msg.SubId {
						state.mySubreddits = append(state.mySubreddits[:i], state.mySubreddits[i+1:]...)
						break
					}
				}
			}

		case *messages.PostResponse:
			metricsMsg := &messages.MetricsMessage{
				Action:       "create_post",
				Success:      msg.Success,
					ResponseTime: time.Since(state.startTime),
					Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)

			if msg.Success {
					state.myPosts = append(state.myPosts, msg.PostId)
					state.userToActorPID["post"] = msg.ActorPID
			}

    case *messages.CreateCommentResponse:
			metricsMsg := &messages.MetricsMessage{
				Action:       "create_comment",
				Success:      msg.Success,
				ResponseTime: time.Since(state.startTime),
				Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)

			if msg.Success {
					state.myComments = append(state.myComments, msg.CommentId)
					state.userToActorPID["comment"] = msg.ActorPID
			}

		case *messages.SendDirectMessageResponse:
			metricsMsg := &messages.MetricsMessage{
				Action:       "send_dm",
				Success:      msg.Success,
				ResponseTime: time.Since(state.startTime),
				Error:        msg.Error,
			}
			context.Send(state.controllerPID, metricsMsg)

			if msg.Success {
					state.myDms = append(state.myDms, msg.MessageID)
					state.userToActorPID["direct_message"] = msg.ActorPID
			}

		// Add this new case in the Receive method
		case *messages.GetSubredditsResponse:
			if msg.Success && len(msg.Subreddits) > 0 {
				// Filter out subreddits the user is already part of
				availableSubreddits := make([]string, 0)
				for _, sub := range msg.Subreddits {
					if !contains(state.mySubreddits, sub) {
						availableSubreddits = append(availableSubreddits, sub)
					}
				}

				if len(availableSubreddits) > 0 {
						// Pick a random subreddit from available ones
							randomSub := availableSubreddits[state.rand.Intn(len(availableSubreddits))]

							joinMsg := &messages.JoinSubreddit{
									SubredditName: randomSub,
									UserId:        state.username,
									ActorPID:      state.userToActorPID["subreddit"],
							}
							context.Request(state.enginePID, joinMsg)
						}
					}
    }
}

func (state *ClientActor) performRandomAction(context actor.Context) {
    action := state.rand.Intn(8)

    switch action {
    case 0:
        state.createSubreddit(context)
    case 1:
        state.createPost(context)
    case 2:
        state.createComment(context)
    case 3:
        state.sendDirectMessage(context)
    case 4:
        state.voteOnPost(context)
    case 5:
        state.voteOnComment(context)
		case 6:
				state.joinRandomSubreddit(context)
		case 7:
				state.leaveRandomSubreddit(context)
    }
}

// Individual actions
func (state *ClientActor) register(context actor.Context) {
		state.startTime = time.Now()
    msg := &messages.RegisterUser{
        Username: fmt.Sprintf("user_%d", state.rand.Intn(100000)),
        Password: fmt.Sprintf("pass_%d", state.rand.Intn(100000)),
    }
    context.Request(state.enginePID, msg)
}

func (state *ClientActor) login(context actor.Context) {
    msg := &messages.LoginUser{
        Username: state.username,
        Password: state.username, // Simplified for demo
        ActorPID: state.userToActorPID["user"],
    }
    context.Request(state.enginePID, msg)
}

func (state *ClientActor) createSubreddit(context actor.Context) {
    msg := &messages.CreateSubreddit{
        Name:        fmt.Sprintf("subreddit_%d", state.rand.Intn(10000)),
        Description: state.generateContent(),
        CreatorId:   state.username,
				ActorPID: state.userToActorPID["subreddit"],
    }
    context.Request(state.enginePID, msg)
}

func (state *ClientActor) joinRandomSubreddit(context actor.Context) {
	msg := &messages.GetSubreddits{
		ActorPID: state.userToActorPID["subreddit"],
	}

	context.Request(state.enginePID, msg)
}

func (state *ClientActor) leaveRandomSubreddit(context actor.Context) {
	if len(state.mySubreddits) == 0 {
		return
	}

	subredditName := state.mySubreddits[state.rand.Intn(len(state.mySubreddits))]

	msg := &messages.LeaveSubreddit{
			SubredditName: subredditName,
			UserId:        state.username,
			ActorPID:      state.userToActorPID["subreddit"],
	}
	context.Request(state.enginePID, msg)
}

func (state *ClientActor) createPost(context actor.Context) {
    if len(state.mySubreddits) == 0 {
        return
    }

    subreddit := state.mySubreddits[state.rand.Intn(len(state.mySubreddits))]
    msg := &messages.Post{
        Title:         fmt.Sprintf("post_%d", state.rand.Intn(10000)),
        Content:       state.generateContent(),
        AuthorId:      state.username,
        SubredditName: subreddit,
        ActorPID:      state.userToActorPID["post"],
    }

    state.startTime = time.Now() // Set start time for metrics
    response, err := context.RequestFuture(state.enginePID, msg, 5*time.Second).Result()
    if err != nil {
        return
    }

    if postResponse, ok := response.(*messages.PostResponse); ok {
        metricsMsg := &messages.MetricsMessage{
            Action:       "create_post",
            Success:      postResponse.Success,
            ResponseTime: time.Since(state.startTime),
            Error:        postResponse.Error,
        }
        context.Send(state.controllerPID, metricsMsg)

        if postResponse.Success {
            state.myPosts = append(state.myPosts, postResponse.PostId)
        }
    }
}

func (state *ClientActor) createComment(context actor.Context) {
    if len(state.myPosts) == 0 {
        return
    }

    postID := state.myPosts[state.rand.Intn(len(state.myPosts))]
    msg := &messages.CreateComment{
        PostId:   postID,
        ParentId: "",  // Empty for top-level comment
        Content:  state.generateContent(),
        AuthorId: state.username,
        ActorPID: state.userToActorPID["comment"],
    }
    context.Request(state.enginePID, msg)
}

func (state *ClientActor) sendDirectMessage(context actor.Context) {
    msg := &messages.SendDirectMessage{
        FromUserID: state.username,
        ToUserID:   fmt.Sprintf("user_%d", state.rand.Intn(1000)), // Random user
        Content:    state.generateContent(),
				ActorPID:   state.userToActorPID["direct_message"],
    }
    context.Request(state.enginePID, msg)
}

func (state *ClientActor) voteOnPost(context actor.Context) {
    if len(state.myPosts) == 0 {
        return
    }

    postID := state.myPosts[state.rand.Intn(len(state.myPosts))]
    msg := &messages.Vote{
        UserID:    state.username,
        TargetID:  postID,
        IsUpvote:  state.rand.Float32() > 0.5,
				Type:      "post",
				ActorPID:  state.userToActorPID["post"],
    }
    context.Request(state.enginePID, msg)
}

func (state *ClientActor) voteOnComment(context actor.Context) {
    if len(state.myComments) == 0 {
        return
    }

    commentID := state.myComments[state.rand.Intn(len(state.myComments))]
    msg := &messages.Vote{
        UserID:    state.username,
        TargetID:  commentID,
        IsUpvote:  state.rand.Float32() > 0.5,
				Type:      "comment",
				ActorPID:  state.userToActorPID["comment"],
    }
    context.Request(state.enginePID, msg)
}
