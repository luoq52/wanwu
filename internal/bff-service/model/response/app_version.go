package response

type AppVersionInfo struct {
	Version     string `json:"version"`
	Desc        string `json:"desc"`
	CreatedAt   string `json:"created_at"`
	PublishType string `json:"publish_type"`
}
