package contracts

type Resp struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
}
