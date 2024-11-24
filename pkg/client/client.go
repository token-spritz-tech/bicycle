package client

import (
	"bicycle/response"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

func NewClient(baseUrl string, authKey string) *Client {
	c := &Client{
		BaseUrl: baseUrl,
		Client:  req.NewClient().SetTimeout(5*time.Second).SetCommonHeader("key", authKey),
		AuthKey: authKey,
	}
	return c
}

type Client struct {
	BaseUrl string
	Client  *req.Client
	AuthKey string
}

func queryParams(data interface{}) (map[string]string, error) {
	result := make(map[string]string)
	val := reflect.ValueOf(data)
	typ := val.Type()

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only accepts structs; got %T", data)
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		valueField := val.Field(i)

		// 跳过非导出的字段
		if !valueField.CanInterface() {
			continue
		}

		tag, ok := field.Tag.Lookup("json")
		if !ok || tag == "-" {
			continue // 没有 json 标签或标签是 "-" 忽略
		}

		key := strings.SplitN(tag, ",", 2)[0] // 取逗号前的部分作为key
		if key == "" {
			key = field.Name // 如果 json 标签为空, 则使用字段名
		}

		var value string
		switch valueField.Kind() {
		case reflect.String:
			value = valueField.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = strconv.FormatInt(valueField.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value = strconv.FormatUint(valueField.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			value = strconv.FormatFloat(valueField.Float(), 'f', -1, 64)
		case reflect.Bool:
			value = strconv.FormatBool(valueField.Bool())
		default:
			// 对于复杂的情况，我们暂时将其 JSON 编码
			jsonBytes, err := json.Marshal(valueField.Interface())
			if err != nil {
				return nil, fmt.Errorf("error encoding %s - %v", key, err)
			}
			value = string(jsonBytes)
		}

		result[key] = value
	}

	return result, nil
}

func (s *Client) createRequest(ctx context.Context) *req.Request {
	return s.Client.R().SetContext(ctx)
}

func (s *Client) DevMode() *Client {
	s.Client.DevMode()
	return s
}

type Response struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (s *Client) ParseResponse(resp *req.Response, result any) (err error) {
	reply := &Response{Data: result}
	if err := resp.Unmarshal(reply); err != nil {
		return err
	}
	if reply.Code != 0 {
		return response.NewServiceError(reply.Code, reply.Msg)()
	}
	return
}

func (s *Client) Get(ctx context.Context, url string, req any, result any) (err error) {
	params, err := queryParams(req)
	if err != nil {
		return
	}

	resp, err := s.createRequest(ctx).SetQueryParams(params).Get(url)
	if err != nil {
		return
	}

	return s.ParseResponse(resp, result)
}

func (s *Client) Post(ctx context.Context, url string, data any, result any) (err error) {
	resp, err := s.createRequest(ctx).SetBody(data).Post(url)
	if err != nil {
		return
	}

	return s.ParseResponse(resp, result)
}
