package logic

import (
	"context"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSensitiveWordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSensitiveWordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSensitiveWordsLogic {
	return &ListSensitiveWordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSensitiveWordsLogic) ListSensitiveWords(in *risk.ListSensitiveWordsReq) (*risk.ListSensitiveWordsResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.PageSize
	if size <= 0 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	where := "status=1"
	args := []any{}
	if in.Category != "" {
		where += " AND category=?"
		args = append(args, in.Category)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM sensitive_word WHERE "+where, args...); err != nil {
		return nil, err
	}

	listArgs := append(args, size, offset)
	var rows []*sensitiveWordRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+sensitiveWordCols+" FROM sensitive_word WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", listArgs...); err != nil {
		return nil, err
	}

	out := make([]*risk.SensitiveWord, 0, len(rows))
	for _, r := range rows {
		out = append(out, toSensitiveWordProto(r))
	}
	return &risk.ListSensitiveWordsResp{Words: out, Total: total}, nil
}
