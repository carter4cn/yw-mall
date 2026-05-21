// mall-user-rpc/internal/logic/emailhelpers.go
package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// EmailSend mock — 把 (target, code) 打到 logx，生产替换成 SES/SMTP 实现。
// 跟 S4.1 SmsSend 同模式，方便统一升级。
func EmailSend(ctx context.Context, target, code string) error {
	logx.WithContext(ctx).Infof("[mock-email] target=%s code=%s", target, code)
	return nil
}
