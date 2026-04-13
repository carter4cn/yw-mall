package result

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"mall-common/errorx"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(w http.ResponseWriter, data interface{}) {
	httpx.OkJsonCtx(nil, w, &Response{
		Code: errorx.OK,
		Msg:  "success",
		Data: data,
	})
}

func Fail(w http.ResponseWriter, err error) {
	if codeErr, ok := err.(*errorx.CodeError); ok {
		httpx.OkJsonCtx(nil, w, &Response{
			Code: codeErr.Code,
			Msg:  codeErr.Msg,
		})
	} else {
		httpx.OkJsonCtx(nil, w, &Response{
			Code: errorx.ServerError,
			Msg:  err.Error(),
		})
	}
}
