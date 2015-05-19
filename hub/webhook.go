package hub

type (
	Webhook struct {
		PushData    *PushData   `json:"push_data,omitempty"`
		Repository  *Repository `json:"repository,omitempty"`
		CallbackUrl string      `json:"callback_url,omitempty"`
	}
)
