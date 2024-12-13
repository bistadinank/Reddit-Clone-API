package messages

import "github.com/asynkron/protoactor-go/actor"

type CreateComment struct {
    PostId      string
    ParentId    string  // Empty for top-level comments, CommentId for replies
    Content     string
    AuthorId    string
    ActorPID    *actor.PID
}

type CreateCommentResponse struct {
    Success   bool
    Error     string
    CommentId string
    ActorPID  *actor.PID
}

type Comment struct {
    CommentId  string
    PostId     string
    ParentId   string
    Content    string
    AuthorId   string
    Replies    []*Comment  // Nested replies
    ActorPID   *actor.PID
    Timestamp  int64
    VoteCount  int        // Add this field
}

type ListPostComments struct {
    PostId   string
    ActorPID *actor.PID
}

type ListPostCommentsResponse struct {
    Success  bool
    Error    string
    Comments []*Comment  // Will contain nested structure
}

type EditComment struct {
    CommentId string
    Content   string
    AuthorId  string    // To verify ownership
    ActorPID  *actor.PID
}

type EditCommentResponse struct {
    Success  bool
    Error    string
    ActorPID *actor.PID
}

type DeleteComment struct {
    CommentId string
    AuthorId  string    // To verify ownership
    ActorPID  *actor.PID
}

type DeleteCommentResponse struct {
    Success  bool
    Error    string
    ActorPID *actor.PID
}
