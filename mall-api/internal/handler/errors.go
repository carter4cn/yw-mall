package handler

import (
	"context"
	"net/http"
	"strings"

	"mall-common/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// writeBizError 把 gRPC error 解包成 {code, message} JSON 返给前端。
// 配合 FE request.ts 在非 2xx 时读 body.message — 前端能看到真实业务消息
// 而不是 fallback 的 "请求失败"。
//
// 默认 400；NotFound→404，Unauthenticated→401，Internal→500。
// 所有业务 logic 用 errors.New(...) 返的错（gRPC code=Unknown）一律映射 400。
func writeBizError(ctx context.Context, w http.ResponseWriter, err error) {
	msg := unwrapGrpcDesc(err)
	httpStatus := http.StatusBadRequest
	bizCode := errorx.ParamError

	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.NotFound:
			httpStatus = http.StatusNotFound
			bizCode = errorx.NotFound
		case codes.Unauthenticated:
			httpStatus = http.StatusUnauthorized
			bizCode = errorx.AuthError
		case codes.Internal:
			httpStatus = http.StatusInternalServerError
			bizCode = errorx.ServerError
		}
	}

	httpx.WriteJsonCtx(ctx, w, httpStatus, map[string]any{
		"code":    bizCode,
		"message": msg,
	})
}

// unwrapGrpcDesc 提取 gRPC error 的 desc 部分。
// status.FromError 拿不到时回退正则抠 "desc = "。
func unwrapGrpcDesc(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	msg := err.Error()
	if i := strings.Index(msg, "desc = "); i >= 0 {
		return msg[i+len("desc = "):]
	}
	return msg
}
