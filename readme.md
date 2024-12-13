
# Reddit Clone - Distributed Actor System

### Team Members

-   Anup Sedhain (UF ID - 92896347)
-   Dinank Bista (UF ID - 41975568)

## Project Overview

This project implements a distributed Reddit-like platform using the Actor model in Go with Proto.Actor framework. The system simulates a social platform where users can create subreddits, post content, comment, vote, and send direct messages.

## Architecture

### Core Components

#### 1. Actor System

The system is built on several specialized actors:

```
// Main Actor Types
- ClientActor     // Simulates user behavior
- EngineActor    // Coordinates system operations
- PostActor      // Manages posts and voting
- SubredditActor // Handles subreddit operations
- CommentActor   // Manages comments
```

#### 2. Controller

The simulation is managed by a controller that:

-   Creates and manages 1000 client actors
-   Creates 10 engine actors for load balancing
-   Collects and reports metrics

```
func (sc *SimulationController) StartSimulation() error {
    // Create engine actors
    for i := 0; i < 10; i++ {
        engineProps := actor.PropsFromProducer(func() actor.Actor {
            return actors.NewEngineActor()
        })
        enginePID := sc.system.Root.Spawn(engineProps)
        sc.enginePIDs = append(sc.enginePIDs, enginePID)
    }

   // Create client actors
   for i := 0; i < 1000; i++ {
       clientProps := actor.PropsFromProducer(func() actor.Actor {
           return actors.NewClientActor(sc.getEngineActor(), sc.pid)
       })
       clientPID := sc.system.Root.Spawn(clientProps)
       sc.clientPIDs = append(sc.clientPIDs, clientPID)
   }
   return nil
}
```

## Client and Engine Analysis

These two actor systems represent the core distributed architecture of the Reddit clone. Let's see how they work together:

### Architecture Overview

```
ClientActor (Multiple Instances) →→→ EngineActor (Load Balancer) →→→ Service Actors
```

#### 1. Client (The Frontend Process)

```
type  ClientActor  struct  {
enginePID  *actor.PID  // Connection to backend
userToActorPID  map[string]*actor.PID  // Cache of service locations
mySubreddits  []string  // Local state
myPosts  []string
myComments  []string
myDms  []string
actionDelay  time.Duration  // Simulated user think time
}
```

Key Features:

-   Maintains local state of user's activities
-   Simulates realistic user behavior with random actions
-   Implements client-side caching of actor PIDs
-   Handles metrics collection for each operation

#### 2. Engine (The Backend Process)
```
type  EngineActor  struct  {
userActors  []*actor.PID
postActors  []*actor.PID
subredditActors  []*actor.PID
directMessageActors  []*actor.PID
commentActors  []*actor.PID
currentUserActor  int
}
```

Key Features:

-   Acts as a service discovery and load balancer
-   Manages pools of service actors
-   Implements round-robin distribution
-   Handles message routing with actor affinity

### Communication Pattern

-   First Request (without ActorPID)
    ```
    // Client side
    msg :=  &messages.CreatePost{...}
    context.Request(state.enginePID, msg)

    // Engine side
    postActor := state.postActors[state.currentUserActor]
    context.RequestWithCustomSender(postActor, msg, context.Sender())
    ```

-   Subsequent Requests (with cached ActorPID)
    ```
    // Client side
    msg :=  &messages.CreatePost{
    ActorPID: state.userToActorPID["post"]
    ...
    }
    context.Request(state.enginePID, msg)

    // Engine side
    context.RequestWithCustomSender(msg.ActorPID, msg, context.Sender())
    ```

### Optimization Strategies

-   Client-Side
    ```
    userToActorPID map[string]*actor.PID  // Caches service locations
    ```
    - Reduces load on engine actor
-   Enables direct routing to service actors
-   Maintains affinity for better caching

-   Server-Side
    ```
    // Creates 10 instances of each service type
    for i :=  0; i < 10; i++ {
    userProps := actor.PropsFromProducer(func()  actor.Actor  {
    return  NewUserActor()
    }
    ```
    - Horizontal scaling of services
-   Round-robin load balancing
-   Service isolation


## Data Distribution

#### Posts

Posts are managed by the PostActor with three main data structures:

```
type PostActor struct {
    posts map[string]*messages.Post           // PostId -> Post
    subredditPosts map[string][]string       // SubredditName -> []PostId
    votes map[string]map[string]bool         // PostId -> UserId -> IsUpvote
}
```

#### Client State

Each client maintains its own state:

```
type ClientActor struct {
    mySubreddits  []string    // subscribed subreddits
    myPosts       []string    // created posts
    myComments    []string    // created comments
    myDms         []string    // sent direct messages
}
```

### Load Balancing

The system uses Zipf distribution for intelligent load balancing across engine actors:
```
zipf := rand.NewZipf(
    rand.New(rand.NewSource(time.Now().UnixNano())),
    1.1,  // skewness
    1.0,  // value parameter
    10    // number of engines
)
```

## Key Features

### 1. Random Action Simulation

Clients perform random actions to simulate real user behavior:

```
func (state *ClientActor) performRandomAction(context actor.Context) {
    action := state.rand.Intn(8)
    switch action {
    case 0: state.createSubreddit(context)
    case 1: state.createPost(context)
    case 2: state.createComment(context)
    case 3: state.sendDirectMessage(context)
    case 4: state.voteOnPost(context)
    case 5: state.voteOnComment(context)
    case 6: state.joinRandomSubreddit(context)
    case 7: state.leaveRandomSubreddit(context)
    }
}
```

### 2. Metrics Collection

The system tracks comprehensive metrics:

```
type SimulationMetrics struct {
    TotalRequests      int64
    SuccessfulRequests int64
    FailedRequests     int64
    ActionCounts       map[messages.ActionType]int64
    ResponseTimes      map[messages.ActionType][]time.Duration
    ErrorCounts        map[messages.ActionType]int64
}
```

## Running the System

1. Start the simulation:

```
# Run with custom values

go  run  main.go  -clients  1000  -engines  10  -duration  10
```

2. The system will:

-   Initialize actors
-   Run for "duration" seconds
-   Print detailed metrics

```
go run main.go -clients 5000 -engines 10 -duration 5
2:36PM INF actor system started lib=Proto.Actor system=Nk89G54LwEL3X3REE6jgM7 id=Nk89G54LwEL3X3REE6jgM7
Starting simulation...

Simulation Results:
Total Requests: 110139
Successful Requests: 106333
Failed Requests: 3806

Average Response Times:
create_subreddit: 2.51203271s
send_dm: 2.512891688s
join_subreddit: 2.539153009s
create_post: 2.687613563s
leave_subreddit: 2.716639299s
create_comment: 2.739770432s
register: 7.799µs
login: 13.801µs

Action Counts:
create_subreddit: 19260
send_dm: 18988
join_subreddit: 18459
create_post: 15288
leave_subreddit: 14685
create_comment: 13464
register: 5000
login: 4996

Error Counts:
register: 4
join_subreddit: 3301
create_subreddit: 479
leave_subreddit: 22
```

## Comparision between two simulations

### Scale Comparison
Simulation 1: 10,000 clients, 10 engines, 10 seconds
Simulation 2: 5,000 clients, 10 engines, 5 seconds

### 1. Request Volume & Success Rates

Simulation 1:

- Total: 528,196 (~52.8K/sec)

- Success Rate: 96.6% (510,255/528,196)

- Failure Rate: 3.4% (17,941/528,196)

Simulation 2:

- Total: 110,139 (~22K/sec)

- Success Rate: 96.5% (106,333/110,139)

- Failure Rate: 3.5% (3,806/110,139)

Key Insight: Both simulations maintain similar success rates (~96.5%) despite Simulation 1 handling more than double the requests per second, showing good scalability.

### 2. Response Times

Simulation 1 (10K clients):

- Content Operations: ~5-5.3 seconds

- Auth Operations: 6-12 microseconds

Simulation 2 (5K clients):

- Content Operations: ~2.5-2.7 seconds

- Auth Operations: 7-13 microseconds

Key Insights:

-   Response times roughly doubled with double the clients

-   Authentication operations (register/login) remain consistently fast in microseconds

-   Content operations show linear scaling with load

### 3. Action Distribution

Simulation 1 (Most to Least):

1. send_dm: 91,585

2. create_subreddit: 91,364

3. join_subreddit: 90,909

4. leave_subreddit: 80,246

5. create_post: 80,217

6. create_comment: 73,904

Simulation 2 (Most to Least):

1. create_subreddit: 19,260

2. send_dm: 18,988

3. join_subreddit: 18,459

4. create_post: 15,288

5. leave_subreddit: 14,685

6. create_comment: 13,464

Key Insight: Action distribution patterns remain relatively consistent between simulations, showing stable behavior patterns regardless of scale.

### 4. Error Analysis

Simulation 1 Errors:

- create_subreddit: 8,800

- join_subreddit: 8,732

- leave_subreddit: 399

- register: 13

Simulation 2 Errors:

- join_subreddit: 3,301

- create_subreddit: 479

- leave_subreddit: 22

- register: 4

Key Insights:

-   Subreddit operations (create/join) are the most error-prone in both simulations

-   Registration remains highly reliable even at scale

-   Error patterns are consistent but magnified with scale

### Overall Conclusions:

1. Scalability

-   System shows linear scaling in response times

-   Maintains consistent success rates across different loads

-   Authentication system performs exceptionally well at both scales

2. Performance Bottlenecks

-   Subreddit operations are the main source of errors

-   Content operations show significant latency compared to auth operations

-   Response times scale linearly with load, suggesting no catastrophic degradation

3. System Behavior

-   Action patterns remain consistent across scales

-   Error distributions are predictable

-   Authentication remains fast and reliable

## Increasing engine number

Here are the most interesting findings when comparing the large-scale simulation (50K clients, 100 engines) with the previous two:

### 1. Engine Scaling Impact

10K clients, 10 engines: ~5.2s response time

50K clients, 100 engines: ~4.8s response time

Key Insight: Despite 5x more clients, response times slightly improved with 10x more engines, suggesting efficient horizontal scaling.

### 2. Error Rate Evolution

5K clients: 3.5% failure rate

10K clients: 3.4% failure rate

50K clients: 4.3% failure rate

Interesting: Error rate remains relatively stable even with 10x scale increase, showing good system resilience.

### 3. Authentication Performance

Register/Login times:

5K clients: 7-13 microseconds

50K clients: 16-27 microseconds

Notable: Auth operations only doubled in latency despite 10x user increase, showing excellent scalability for authentication.

### 4. Critical Finding: Subreddit Operations

Join Subreddit Error Rates:

5K clients: 3,301 errors

50K clients: 41,207 errors

Problem Area: Subreddit operations don't scale linearly with the system - errors increased by ~12x when system scaled by 10x, indicating a potential bottleneck in subreddit management.

### 5. Workload Distribution

The ratio of operations remains consistent across all scales, suggesting the random action generator maintains stable behavior patterns even at larger scales. This validates the simulation's reliability in representing consistent user behavior.

### Key Takeaway

The system shows excellent horizontal scaling capabilities for most operations, but subreddit management emerges as the clear bottleneck that needs architectural attention at larger scales.



## Limitations and Future Improvements

### Current Limitations

1. In-memory data storage only

-   No persistence layer

-   Limited error handling

-   Basic authentication

### Recommended Improvements

#### 1. Data Persistence

-   Add database integration

-   Implement data recovery mechanisms

-   Add caching layer

#### 2. Scaling

-   Implement actor supervision strategies

-   Add cluster support

-   Implement sharding for better data distribution

#### 3. Features

-   Add moderation capabilities

-   Implement content filtering

-   Enhanced user authentication

-   Media content support

#### 4. Monitoring

-   Real-time metrics dashboard

-   Request flow tracing

-   Performance monitoring

#### 5. Resilience

-   Circuit breakers

-   Rate limiting

-   Enhanced error handling

## Error Hypothesis

### Hypotheses for High Error Rates:

1. Race Conditions in Subreddit Creation
- Multiple clients might try to create subreddits with the same name simultaneously
- No apparent locking mechanism in the shared subreddit map
- Could lead to concurrent map writes or duplicate name conflicts

2. Memory Contention
- Large number of concurrent modifications to membership lists
- Possible memory pressure from growing maps without cleanup
- Could lead to resource exhaustion or slow operations

3. Message Queue Overflow
- Zipf distribution might be overloading certain engines
- Popular subreddits might be handled by same engines due to distribution
- Could lead to message queue buildup and timeouts

### Supporting Evidence:

-   Error Rate Scaling
    5K clients: 3,301 join errors
    50K clients: 41,207 join errors
    - Error rate increases non-linearly with scale
-   Suggests concurrent access issues rather than simple capacity problems
-   Operation Timing
    create_subreddit: 4.61s
    join_subreddit: 4.63s
    - Similar timing for both operations
-   Suggests they might be blocking each other

### Potential Solution:

-   Two-Phase Operations

	  1. Check subreddit exists + lock membership

	  2. Perform join operation

	  3. Release lock



## Technical Insights

### Message-Based Communication

The system uses a comprehensive message system for all operations:

```
type Post struct {
    PostId        string
    Title         string
    Content       string
    AuthorId      string
    SubredditName string
    Timestamp     int64
    ActorPID      *actor.PID
}
```

### Voting System

Implements a Reddit-like voting mechanism:

```
func (state *PostActor) handleVote(msg *messages.Vote) *messages.VoteResponse {
    if _, exists := state.posts[msg.TargetID]; !exists {
        return &messages.VoteResponse{Success: false, Error: "Post not found"}
    }

    if state.votes[msg.TargetID] == nil {
        state.votes[msg.TargetID] = make(map[string]bool)
    }

    previousVote, hasVoted := state.votes[msg.TargetID][msg.UserID]
    if hasVoted {
        if previousVote == msg.IsUpvote {
            delete(state.votes[msg.TargetID], msg.UserID)
        } else {
            state.votes[msg.TargetID][msg.UserID] = msg.IsUpvote
        }
    } else {
        state.votes[msg.TargetID][msg.UserID] = msg.IsUpvote
    }

    return &messages.VoteResponse{Success: true}
}
```

## Conclusion

This project demonstrates a practical implementation of a distributed system using the Actor model. Its modular design and message-passing architecture provide a solid foundation for a scalable social platform. While there are areas for improvement, the current implementation successfully showcases key concepts of distributed systems and concurrent programming.

The use of Proto.Actor provides robust actor management and message passing, while the custom metrics collection offers valuable insights into system performance. The project serves as an excellent example of how to structure and implement a distributed system using modern Go practices.
