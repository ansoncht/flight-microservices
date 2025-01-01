package poster

type Poster interface {
	CreatePost(content string, media []string) (string, error)
	PublishPost(postID string) (bool, error)
}
