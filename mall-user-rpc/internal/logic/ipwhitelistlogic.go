package logic

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type IpWhitelistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIpWhitelistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IpWhitelistLogic {
	return &IpWhitelistLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

type ipWhitelistRow struct {
	Id         uint64 `db:"id"`
	AdminId    uint64 `db:"admin_id"`
	Cidr       string `db:"cidr"`
	Note       string `db:"note"`
	CreateTime int64  `db:"create_time"`
}

func (l *IpWhitelistLogic) List(in *user.ListAdminIpWhitelistReq) (*user.ListAdminIpWhitelistResp, error) {
	if in.AdminId <= 0 {
		return nil, errors.New("admin_id required")
	}
	var rows []*ipWhitelistRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, admin_id, cidr, note, create_time FROM admin_ip_whitelist WHERE admin_id=? ORDER BY id DESC",
		in.AdminId); err != nil {
		return nil, err
	}
	out := make([]*user.AdminIpWhitelistEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, &user.AdminIpWhitelistEntry{
			Id:         int64(r.Id),
			AdminId:    int64(r.AdminId),
			Cidr:       r.Cidr,
			Note:       r.Note,
			CreateTime: r.CreateTime,
		})
	}
	return &user.ListAdminIpWhitelistResp{Items: out}, nil
}

func (l *IpWhitelistLogic) Add(in *user.AddAdminIpWhitelistReq) (*user.AddAdminIpWhitelistResp, error) {
	if in.AdminId <= 0 {
		return nil, errors.New("admin_id required")
	}
	cidr := strings.TrimSpace(in.Cidr)
	if cidr == "" {
		return nil, errors.New("cidr required")
	}
	// Single IP shorthand: "1.2.3.4" → "1.2.3.4/32"
	if !strings.Contains(cidr, "/") {
		ip := net.ParseIP(cidr)
		if ip == nil {
			return nil, errors.New("CIDR/IP 格式错误")
		}
		if ip.To4() != nil {
			cidr += "/32"
		} else {
			cidr += "/128"
		}
	}
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return nil, errors.New("CIDR 格式错误: " + err.Error())
	}
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"INSERT INTO admin_ip_whitelist (admin_id, cidr, note, create_time) VALUES (?, ?, ?, ?)",
		in.AdminId, cidr, strings.TrimSpace(in.Note), time.Now().Unix())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &user.AddAdminIpWhitelistResp{Id: id}, nil
}

func (l *IpWhitelistLogic) Delete(in *user.DeleteAdminIpWhitelistReq) (*user.OkResp, error) {
	if in.Id <= 0 || in.AdminId <= 0 {
		return nil, errors.New("id and admin_id required")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"DELETE FROM admin_ip_whitelist WHERE id=? AND admin_id=?",
		in.Id, in.AdminId); err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}

// CheckAdminIPAllowed returns nil when the IP either matches an existing
// whitelist row OR the whitelist is empty (= no restriction). Used at admin
// gateway login time after the password check, before minting a session.
//
// Lives here so the matching logic stays next to the CRUD code; the gateway
// only sees the boolean result via the existing ListAdminIpWhitelist RPC.
func CheckAdminIPAllowed(rows []*user.AdminIpWhitelistEntry, ip string) bool {
	if len(rows) == 0 {
		return true
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	for _, r := range rows {
		_, ipNet, err := net.ParseCIDR(r.Cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(parsed) {
			return true
		}
	}
	return false
}
