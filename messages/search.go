package messages

type SearchPosts struct {
    Query string
}

type SearchPostsResponse struct {
    Success bool
    Error   string
    Posts   []*PostFeed
} 