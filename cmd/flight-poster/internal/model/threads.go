package model

// ThreadsUserResponse represents the structure of user data returned by the Threads API.
type ThreadsUserResponse struct {
	ID string `json:"id"` // User ID for the Threads account
}

// ThreadsPostCreationResponse represents the structure of post data returned by the Threads API.
type ThreadsPostCreationResponse struct {
	ID string `json:"id"` // Container ID for the Threads post
}
