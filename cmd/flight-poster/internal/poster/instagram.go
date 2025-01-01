package poster

import "log/slog"

type InstagramPoster struct {
	userID string
	token  string
}

func (ip *InstagramPoster) GetToken() string {
	slog.Info("Getting Token", "user", ip.userID)
	return ip.token
}
