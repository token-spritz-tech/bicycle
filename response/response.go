package response

type Response struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (resp *Response) SetData(data interface{}) *Response {
	resp.Data = data
	return resp
}

func (resp *Response) SetMsg(msg string) *Response {
	resp.Msg = msg
	return resp
}

func (resp *Response) SetCode(code int32) *Response {
	resp.Code = code
	return resp
}
