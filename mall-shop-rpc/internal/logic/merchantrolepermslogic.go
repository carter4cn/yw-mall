package logic

// M1 商家工作台：4 个固定角色 + 权限码映射。
// owner 用 "*" 占位表示"全开"，hasPerm 时特判即可。
//
// 权限码格式：<domain>.<action>，与 admin-api 路由 meta.perms 对齐。

const (
	RoleOwner     = "owner"
	RoleService   = "service"
	RoleWarehouse = "warehouse"
	RoleFinance   = "finance"
)

// RolePerms 各角色的权限码集合。
var RolePerms = map[string][]string{
	RoleOwner: {"*"},
	RoleService: {
		"shop.read", "product.read",
		"order.read", "order.write",
		"refund.read", "refund.handle",
		"staff.read",
	},
	RoleWarehouse: {
		"shop.read", "product.read", "product.write",
		"order.read", "order.ship",
		"refund.read", "refund.inspect",
		"staff.read", "freight.read",
	},
	RoleFinance: {
		"shop.read", "order.read",
		"finance.read", "finance.write",
		"staff.read",
	},
}

// IsValidRole 校验 role 字符串是否在 4 个预定义角色中。空串非法。
func IsValidRole(r string) bool {
	if r == "" {
		return false
	}
	_, ok := RolePerms[r]
	return ok
}
