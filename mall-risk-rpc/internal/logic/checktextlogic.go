package logic

import (
	"context"

	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckTextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckTextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTextLogic {
	return &CheckTextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckText scans the input against all active sensitive words.
// verdict: "clean" | "flag" (any flag matches) | "block" (any block match wins).
func (l *CheckTextLogic) CheckText(in *risk.CheckTextReq) (*risk.CheckTextResp, error) {
	if in.Text == "" {
		return &risk.CheckTextResp{Clean: true, Verdict: "clean"}, nil
	}
	words, err := sensitiveCache.loadAllActive(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	matches := scanText(in.Text, words)
	if len(matches) == 0 {
		return &risk.CheckTextResp{Clean: true, Verdict: "clean"}, nil
	}
	verdict := "flag"
	out := make([]*risk.SensitiveWord, 0, len(matches))
	for _, m := range matches {
		if m.Action == "block" {
			verdict = "block"
		}
		out = append(out, toSensitiveWordProto(m))
	}
	return &risk.CheckTextResp{Clean: false, Verdict: verdict, Matches: out}, nil
}
