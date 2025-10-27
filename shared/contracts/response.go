package contracts

type Resp struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       any
}
