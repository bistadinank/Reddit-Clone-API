# Reddit Clone - Distributed Actor System with REST API

## Team Members
- Anup Sedhain (UF ID - 92896347)
- Dinank Bista (UF ID - 41975568)

## Project Overview
This project implements a distributed Reddit-like platform combining two key architectural components:
1. A REST API interface for client-server communication
2. An Actor-based backend system using Proto.Actor framework in Go

The system enables users to create communities (subreddits), share content, and interact through comments and votes, all while maintaining data consistency across concurrent operations.

## Architecture

### Core Components

#### 1. Actor System
The backend is built on several specialized actors:
```
- EngineActor    // Coordinates system operations
- PostActor      // Manages posts and voting
- SubredditActor // Handles subreddit operations
- CommentActor   // Manages comments and replies
- UserActor      // Handles user operations
```

#### 2. REST API Layer
The system exposes RESTful endpoints that follow Reddit-like API patterns:
```
Authentication:
- POST /register
- POST /login

Content:
- POST /post
- GET  /post/:postId
- POST /comment
- GET  /feed

Communities:
- POST /subreddit
- GET  /subreddits
- POST /subreddit/:name/join
```

### Data Management

#### 1. Post Management
```go
type PostActor struct {
    posts map[string]*messages.Post           // PostId -> Post
    subredditPosts map[string][]string       // SubredditName -> []PostId
    votes map[string]map[string]bool         // PostId -> UserId -> IsUpvote
}
```

#### 2. User State
```go
type UserActor struct {
    karma map[string]int           // UserId -> Karma
    joinedSubreddits map[string][]string  // UserId -> []SubredditName
}
```

## Implementation Details

### 1. REST API Implementation

#### Authentication Flow
1. Registration
   - Username/password validation
   - Account creation
   - Initial state setup

2. Login
   - Credential verification
   - JWT token generation
   - Session management

#### Content Operations
1. Post Creation
   - Authorization check
   - Content validation
   - Actor message dispatch
   - Response handling

2. Comment Management
   - Nested comment support
   - Vote tracking
   - Real-time updates

### 2. Actor System Implementation

#### Message Flow
```
Client Request -> REST Handler -> Engine Actor -> Specialized Actor -> Response
```

#### State Management
- Isolated actor states
- Concurrent access control
- Message-based updates
- Real-time consistency

## Testing Framework

### 1. Test Scenarios
We validate the system using three distinct user personas:

#### Techie (Technical User)
- Creates programming content
- Participates in technical discussions
- Tests post creation and editing

#### Gamer (Gaming Enthusiast)
- Manages gaming communities
- Creates gaming-related content
- Tests cross-posting capabilities

#### MovieBuff (Entertainment Fan)
- Manages movie discussions
- Tests content interaction
- Demonstrates multi-subreddit usage

### 2. Test Flow
```
1. User Creation and Authentication
2. Subreddit Creation and Joining
3. Content Creation and Interaction
4. Search and Feed Generation
5. Edit and Delete Operations
```

## API Endpoints Detail

### Authentication
```
POST /register
- Request: {username, password}
- Response: {token, userId}

POST /login
- Request: {username, password}
- Response: {token}
```

### Content Management
```
POST /post
- Auth: Required
- Request: {title, content, subredditName}
- Response: {postId, success}

POST /comment
- Auth: Required
- Request: {postId, content}
- Response: {commentId, success}
```

### Community Operations
```
POST /subreddit
- Auth: Required
- Request: {name, description}
- Response: {subredditId, success}

POST /subreddit/:name/join
- Auth: Required
- Response: {success}
```

### Discovery
```
GET /feed
- Auth: Required
- Response: {posts[], subreddits[]}

GET /search?q=query
- Auth: Required
- Response: {posts[], relevance}
```

## Performance Analysis

### 1. Metrics
- Average response time: 2.5s for content operations
- Authentication operations: ~10μs
- Search operations: 2.7s
- Concurrent user handling: 96.5% success rate

### 2. Load Testing Results
- Successful requests: 96.5%
- Failed requests: 3.5%
- Error distribution analysis
- Concurrent user capacity

## Technical Insights

### 1. Message-Based Communication
The system implements comprehensive message passing:
```go
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

### 2. Voting System
Reddit-like voting mechanism:
```go
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

## Running the System

### Prerequisites
- Go 1.16 or higher
- Proto.Actor framework
- Git

### How to run this program
1. Build and run the main server:
   ```bash
   go build -o reddit
   ./reddit
   ```
   This will start the server on port 8080

2. Run the test script:
   ```bash
   go run script/test_script.go
   ```
   This will execute the test scenarios with different user personas

### Setup Instructions
1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Start the server: `go run main.go`
4. Run tests: `go run test_script.go`

## Future Enhancements

### 1. Security
- Public key-based digital signatures
- Enhanced authentication
- Rate limiting
- Content verification

### 2. Features
- Media content support
- Real-time notifications
- Enhanced search capabilities
- Content recommendation system

### 3. Performance
- Database integration
- Caching layer
- Load balancing improvements
- Response time optimization

## Conclusion
This project demonstrates a practical implementation of a distributed system using the Actor model with REST API integration. Its modular design and message-passing architecture provide a solid foundation for a scalable social platform. The implementation successfully showcases key concepts of distributed systems and concurrent programming while maintaining clean REST principles.


### How to run this program
1. Build and run the main server:
   ```bash
   go build -o reddit
   ./reddit
   ```
   This will start the server on port 8080

2. Run the test script:
   ```bash
   go run script/test_script.go
   ```
   This will execute the test scenarios with different user personas

### Setup Instructions
1. Clone the repository
2. Install dependencies: `go mod tidy`


[Demo Video Link](your_video_link)