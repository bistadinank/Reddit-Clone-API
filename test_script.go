package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

type TestClient struct {
	client  *http.Client
	token   string
}

func NewTestClient() *TestClient {
	return &TestClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (tc *TestClient) makeRequest(method, endpoint string, body interface{}, needsAuth bool) (map[string]interface{}, error) {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if needsAuth {
		req.Header.Set("Authorization", "Bearer "+tc.token)
	}

	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Raw response from server: %s\n", string(respBody))

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		fmt.Printf("Response that caused error: %s\n", string(respBody))
		return nil, err
	}

	return result, nil
}

func runTests() {
	// 1. User Setup
	fmt.Println("\n=== 1. Creating and Logging in Users ===")
	techie := NewTestClient()
	gamer := NewTestClient()
	moviebuff := NewTestClient()

	// Register users
	techie.makeRequest("POST", "/register", map[string]string{
		"username": "techie",
		"password": "pass123",
	}, false)
	
	gamer.makeRequest("POST", "/register", map[string]string{
		"username": "gamer",
		"password": "pass123",
	}, false)
	
	moviebuff.makeRequest("POST", "/register", map[string]string{
		"username": "moviebuff",
		"password": "pass123",
	}, false)

	// Login users
	resp, _ := techie.makeRequest("POST", "/login", map[string]string{
		"username": "techie",
		"password": "pass123",
	}, false)
	techie.token = resp["token"].(string)

	resp, _ = gamer.makeRequest("POST", "/login", map[string]string{
		"username": "gamer",
		"password": "pass123",
	}, false)
	gamer.token = resp["token"].(string)

	resp, _ = moviebuff.makeRequest("POST", "/login", map[string]string{
		"username": "moviebuff",
		"password": "pass123",
	}, false)
	moviebuff.token = resp["token"].(string)

	// 2. Subreddit Creation & Membership
	fmt.Println("\n=== 2. Creating and Joining Subreddits ===")
	techie.makeRequest("POST", "/subreddit", map[string]string{
		"name": "programming",
		"description": "Programming discussions",
	}, true)

	gamer.makeRequest("POST", "/subreddit", map[string]string{
		"name": "gaming",
		"description": "Gaming discussions",
	}, true)

	moviebuff.makeRequest("POST", "/subreddit", map[string]string{
		"name": "movies",
		"description": "Movie discussions",
	}, true)

	// Join subreddits
	fmt.Println("Users joining subreddits...")
	techie.makeRequest("POST", "/subreddit/gaming/join", nil, true)
	gamer.makeRequest("POST", "/subreddit/programming/join", nil, true)
	moviebuff.makeRequest("POST", "/subreddit/programming/join", nil, true)
	moviebuff.makeRequest("POST", "/subreddit/gaming/join", nil, true)

	// 3. Content Creation
	fmt.Println("\n=== 3. Creating Posts ===")
	// Programming posts
	techie.makeRequest("POST", "/post", map[string]string{
		"title": "Go vs Python Performance",
		"content": "Comparing Go and Python for backend: Go wins in concurrent tasks",
		"subredditName": "programming",
	}, true)

	gamer.makeRequest("POST", "/post", map[string]string{
		"title": "Game Development in Go",
		"content": "Using Go for game server development",
		"subredditName": "programming",
	}, true)

	// Gaming posts
	gamer.makeRequest("POST", "/post", map[string]string{
		"title": "Best Gaming Setups 2024",
		"content": "Top gaming PC builds and peripherals",
		"subredditName": "gaming",
	}, true)

	techie.makeRequest("POST", "/post", map[string]string{
		"title": "Gaming PC Build Guide",
		"content": "How to build a gaming PC for Counter-Strike 2",
		"subredditName": "gaming",
	}, true)

	// Movie posts
	moviebuff.makeRequest("POST", "/post", map[string]string{
		"title": "Top Tech Movies 2024",
		"content": "Best movies about technology and programming",
		"subredditName": "movies",
	}, true)

	moviebuff.makeRequest("POST", "/post", map[string]string{
		"title": "Gaming Movies Review",
		"content": "Reviews of movies based on popular games",
		"subredditName": "movies",
	}, true)

	// 4. Interaction & Engagement
	fmt.Println("\n=== 4. Adding Comments and Votes ===")
	// Comments on programming posts
	resp, _ = gamer.makeRequest("POST", "/comment", map[string]string{
		"postId": "post_programming_Go vs Python Performance",
		"content": "Go's concurrency is amazing!",
	}, true)
	comment1Id := resp["commentId"].(string)

	resp, _ = moviebuff.makeRequest("POST", "/comment", map[string]string{
		"postId": "post_programming_Game Development in Go",
		"content": "Great for game servers!",
	}, true)
	comment2Id := resp["commentId"].(string)

	// Comments on gaming posts
	resp, _ = techie.makeRequest("POST", "/comment", map[string]string{
		"postId": "post_gaming_Best Gaming Setups 2024",
		"content": "RTX 4090 is overkill for most games",
	}, true)
	comment3Id := resp["commentId"].(string)

	// Add voting
	fmt.Println("Adding votes...")
	techie.makeRequest("POST", fmt.Sprintf("/comment/%s/vote", comment1Id), map[string]bool{
		"isUpvote": true,
	}, true)
	moviebuff.makeRequest("POST", fmt.Sprintf("/comment/%s/vote", comment1Id), map[string]bool{
		"isUpvote": true,
	}, true)
	gamer.makeRequest("POST", fmt.Sprintf("/comment/%s/vote", comment2Id), map[string]bool{
		"isUpvote": true,
	}, true)
	moviebuff.makeRequest("POST", fmt.Sprintf("/comment/%s/vote", comment3Id), map[string]bool{
		"isUpvote": false,
	}, true)

	// 5. Feed Testing
	fmt.Println("\n=== 5. Testing Feeds ===")
	fmt.Println("\nTechie's Feed (programming, gaming):")
	resp, _ = techie.makeRequest("GET", "/feed", nil, true)
	prettyPrintJSON(resp)

	fmt.Println("\nGamer's Feed (programming, gaming):")
	resp, _ = gamer.makeRequest("GET", "/feed", nil, true)
	prettyPrintJSON(resp)

	fmt.Println("\nMoviebuff's Feed (all subreddits):")
	resp, _ = moviebuff.makeRequest("GET", "/feed", nil, true)
	prettyPrintJSON(resp)

	// 6. Search Testing
	fmt.Println("\n=== 6. Testing Search ===")
	
	fmt.Println("\nCross-subreddit search for 'Go':")
	resp, _ = techie.makeRequest("GET", "/search?q=Go", nil, true)
	prettyPrintJSON(resp)

	fmt.Println("\nSearch for 'Gaming' (across subreddits):")
	resp, _ = techie.makeRequest("GET", "/search?q=Gaming", nil, true)
	prettyPrintJSON(resp)

	fmt.Println("\nSearch for '2024' (all categories):")
	resp, _ = techie.makeRequest("GET", "/search?q=2024", nil, true)
	prettyPrintJSON(resp)

	fmt.Println("\nCase-insensitive search for 'PYTHON':")
	resp, _ = techie.makeRequest("GET", "/search?q=PYTHON", nil, true)
	prettyPrintJSON(resp)

	// 7. Additional Features
	fmt.Println("\n=== 7. Testing Edit and Delete ===")
	
	// Edit post
	fmt.Println("\nEditing post...")
	resp, err := techie.makeRequest("PATCH", "/post/post_programming_Go vs Python Performance", map[string]string{
		"content": "Updated: Go significantly outperforms Python in concurrent tasks",
	}, true)
	if err != nil {
		fmt.Printf("❌ Error editing post: %v\n", err)
	} else if resp["error"] != nil {
		fmt.Printf("❌ Post edit failed: %v\n", resp["error"])
	} else {
		fmt.Printf("✅ Post edited successfully\n")
	}

	// Edit comment
	fmt.Println("\nEditing comment...")
	resp, err = gamer.makeRequest("PATCH", fmt.Sprintf("/comment/%s", comment1Id), map[string]string{
		"content": "Updated: Go's concurrency model is unmatched!",
	}, true)
	if err != nil {
		fmt.Printf("❌ Error editing comment: %v\n", err)
	} else if resp["error"] != nil {
		fmt.Printf("❌ Comment edit failed: %v\n", resp["error"])
	} else {
		fmt.Printf("✅ Comment edited successfully\n")
	}

	// Delete comment
	techie.makeRequest("DELETE", fmt.Sprintf("/comment/%s", comment3Id), nil, true)

	// Leave and rejoin subreddit
	fmt.Println("\nTesting leave/join subreddit:")
	techie.makeRequest("POST", "/subreddit/gaming/leave", nil, true)
	time.Sleep(1 * time.Second)
	techie.makeRequest("POST", "/subreddit/gaming/join", nil, true)

	// Final feed check after all operations
	fmt.Println("\nFinal Feed Check:")
	resp, _ = techie.makeRequest("GET", "/feed", nil, true)
	prettyPrintJSON(resp)
}

func prettyPrintJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(jsonBytes))
}

func main() {
	fmt.Println("Starting Reddit Clone Feature Test...")
	runTests()
	fmt.Println("\nFeature Test Completed!")
} 