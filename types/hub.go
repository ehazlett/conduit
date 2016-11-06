package types

import "time"

type Webhook struct {
	Repository  Repository `json:"repository"`
	PushData    PushData   `json:"push_data"`
	CallbackURL string     `json:"callback_url"`
}

type PushData struct {
	Images   []string  `json:"images"`
	PushedAt time.Time `json:"pushed_at"`
	Pusher   string    `json:"pusher"`
}

type Repository struct {
	Name           string `json:"name"`
	Owner          string `json:"owner"`
	Namespace      string `json:"namespace"`
	RepositoryName string `json:"repo_name"`
	RepositoryURL  string `json:"repo_url"`
}

type CallbackPayload struct {
	State       string `json:"state"`
	Description string `json:"description"`
	Context     string `json:"context"`
	TargetURL   string `json:"target_url"`
}
