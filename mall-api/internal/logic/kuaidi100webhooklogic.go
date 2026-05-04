package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"mall-api/internal/svc"
	"mall-common/errorx"
	logisticspb "mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type Kuaidi100WebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKuaidi100WebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Kuaidi100WebhookLogic {
	return &Kuaidi100WebhookLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

type kuaidi100Pushed struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	AutoCheck  string `json:"autoCheck"`
	ComOld     string `json:"comOld"`
	ComNew     string `json:"comNew"`
	LastResult struct {
		Message string `json:"message"`
		Nu      string `json:"nu"`
		Ischeck string `json:"ischeck"`
		Com     string `json:"com"`
		State   string `json:"state"`
		Status  string `json:"status"`
		Data    []struct {
			Time     string `json:"time"`
			Ftime    string `json:"ftime"`
			Context  string `json:"context"`
			Location string `json:"location"`
			Status   string `json:"status"`
		} `json:"data"`
	} `json:"lastResult"`
}

// VerifySign is exported for unit tests.
func VerifySign(param, sign, key string) bool {
	h := md5.Sum([]byte(param + key))
	return strings.EqualFold(hex.EncodeToString(h[:]), sign)
}

func (l *Kuaidi100WebhookLogic) Process(param, sign string) (string, error) {
	if !VerifySign(param, sign, l.svcCtx.Config.Kuaidi100.WebhookKey) {
		return "", errorx.NewCodeError(errorx.LogisticsKuaidi100SignInvalid)
	}
	var p kuaidi100Pushed
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return "", errorx.NewCodeError(errorx.ParamError)
	}
	events := make([]*logisticspb.Track, 0, len(p.LastResult.Data))
	for _, d := range p.LastResult.Data {
		t, _ := time.Parse("2006-01-02 15:04:05", d.Time)
		events = append(events, &logisticspb.Track{
			TrackTime:   t.Unix(),
			Location:    d.Location,
			Description: d.Context,
		})
	}
	if _, err := l.svcCtx.LogisticsRpc.IngestWebhookEvents(l.ctx, &logisticspb.IngestWebhookEventsReq{
		Carrier:    p.LastResult.Com,
		TrackingNo: p.LastResult.Nu,
		Events:     events,
	}); err != nil {
		return "", err
	}
	return `{"result":true,"returnCode":"200","message":"success"}`, nil
}
