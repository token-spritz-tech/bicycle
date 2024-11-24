package response

type PageInfo struct {
	Records  interface{} `json:"records"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Count    int         `json:"count"`
}
