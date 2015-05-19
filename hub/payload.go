package hub

type CallbackPayload struct {
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	Context     string `json:"context,omitempty"`
	TargetUrl   string `json:"target_url,omitempty"`
}
