package model

// ThreadsUserResponse holds the user ID information returned by the Threads API.
type ThreadsUserResponse struct {
	ID string `json:"id"`
}

// ThreadsContainerResponse holds the container response from the Threads API.
type ThreadsContainerResponse struct {
	ID string `json:"id"`
}

// ThreadsPostResponse holds the post response from the Threads API.
type ThreadsPostResponse struct {
	ID string `json:"id"`
}
