package social

type Poster interface {
	GetToken() string
	CreatePost(content string, media []string) (string, error)
	PublishPost(postID string) (bool, error)
}
