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

// BizErrorHandler 把 gRPC error / errorx.CodeError / 朴素 errors.New(...) 统一
// 解包成 {code, message} JSON 返给前端。
// 配合 FE request.ts 在非 2xx 时读 body.message — 前端能看到真实业务消息
// 而不是 fallback 的 "请求失败"。
//
// 默认 400；errorx.CodeError 用自带 code；NotFound→404，Unauthenticated→401，
// Internal→500。所有业务 logic 用 errors.New(...) 返的错（gRPC code=Unknown）
// 一律映射 400。
//
// 签名与 httpx.SetErrorHandlerCtx 对齐 —— 注册成全局 handler 后，所有
// httpx.Error/ErrorCtx 调用都会走这里，handler 不必逐个手写。
func BizErrorHandler(ctx context.Context, err error) (int, any) {
	// errorx.CodeError 优先（业务码已自带 code+msg）
	if ce, ok := err.(*errorx.CodeError); ok {
		httpStatus := http.StatusBadRequest
		switch ce.Code {
		case errorx.AuthError:
			httpStatus = http.StatusUnauthorized
		case errorx.NotFound, errorx.UserNotFound, errorx.ProductNotFound,
			errorx.OrderNotFound, errorx.ShopNotFound:
			httpStatus = http.StatusNotFound
		case errorx.ServerError:
			httpStatus = http.StatusInternalServerError
		}
		return httpStatus, map[string]any{"code": ce.Code, "message": ce.Msg}
	}

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

	return httpStatus, map[string]any{"code": bizCode, "message": msg}
}

// writeBizError 是 BizErrorHandler 的写入版，留作老 handler 直接调用的兼容入口。
// 新代码直接用 httpx.ErrorCtx —— 全局已注册 BizErrorHandler。
func writeBizError(ctx context.Context, w http.ResponseWriter, err error) {
	httpStatus, body := BizErrorHandler(ctx, err)
	httpx.WriteJsonCtx(ctx, w, httpStatus, body)
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
