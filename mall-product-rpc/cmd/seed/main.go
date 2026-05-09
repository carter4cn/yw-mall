// Seed binary for mall-product-rpc: creates 40 sample products across 6 shops.
// Idempotent: skips products whose name already exists in the product table.
// Run after mall-shop-rpc/cmd/seed so shop IDs exist.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-product-rpc/product"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	productAddr    = flag.String("product", "127.0.0.1:9001", "product rpc address")
	productDS      = flag.String("ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_product?charset=utf8mb4&parseTime=true&loc=Local", "product MySQL DSN")
	shopDS         = flag.String("shop-ds", "proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_shop?charset=utf8mb4&parseTime=true&loc=Local", "shop MySQL DSN")
	minioBase      = flag.String("minio", "http://127.0.0.1:9000/mall-product", "MinIO base URL for product images")
)

type productSeed struct {
	shopName   string
	name       string
	desc       string
	price      int64 // cents
	stock      int64
	categoryId int64
}

var products = []productSeed{
	// 数码旗舰店 (categoryId=1)
	{"数码旗舰店", "华为 Mate 60 Pro 手机", "麒麟芯片，卫星通话，旗舰性能", 699900, 200, 1},
	{"数码旗舰店", "苹果 MacBook Air M3 笔记本", "轻薄便携，续航强劲，适合日常办公", 999900, 150, 1},
	{"数码旗舰店", "索尼 WH-1000XM5 降噪耳机", "业界领先主动降噪，音质出众", 249900, 300, 1},
	{"数码旗舰店", "iPad Pro 12.9 英寸平板", "M2 芯片，Liquid 视网膜显示屏", 879900, 100, 1},
	{"数码旗舰店", "小米 14 Ultra 手机", "徕卡光学镜头，骁龙8 Gen3，拍照神器", 599900, 250, 1},
	{"数码旗舰店", "戴尔 XPS 15 笔记本", "OLED 触屏，创意专业人士首选", 1299900, 80, 1},
	{"数码旗舰店", "AirPods Pro 2 耳机", "H2 芯片，自适应降噪，MagSafe 充电", 179900, 500, 1},
	{"数码旗舰店", "三星 Galaxy Tab S9 平板", "2K AMOLED 屏，S Pen 随附", 649900, 120, 1},

	// 时尚女装馆 (categoryId=2)
	{"时尚女装馆", "春季碎花连衣裙", "优雅气质，修身设计，适合约会通勤", 29900, 500, 2},
	{"时尚女装馆", "高腰阔腿牛仔裤", "显瘦显腿长，百搭必备单品", 19900, 800, 2},
	{"时尚女装馆", "羊绒混纺长款大衣", "保暖时尚，高端面料，秋冬必备", 89900, 300, 2},
	{"时尚女装馆", "V领雪纺衬衫", "清新飘逸，多色可选，职场休闲两用", 15900, 600, 2},
	{"时尚女装馆", "格纹针织毛衣", "秋冬温暖，复古格纹，百搭显白", 25900, 400, 2},
	{"时尚女装馆", "丝绒吊带小黑裙", "性感优雅，晚宴派对首选", 35900, 200, 2},
	{"时尚女装馆", "运动风休闲套装", "舒适透气，时尚运动，居家外出均可", 22900, 700, 2},

	// 运动户外专营 (categoryId=3)
	{"运动户外专营", "耐克 Air Max 270 跑鞋", "气垫缓震，时尚运动，舒适耐穿", 89900, 400, 3},
	{"运动户外专营", "lululemon 高腰瑜伽裤", "裸感面料，高弹力，塑形显瘦", 79900, 350, 3},
	{"运动户外专营", "迪卡侬登山包 40L", "防水耐磨，多仓储物，户外必备", 39900, 200, 3},
	{"运动户外专营", "Under Armour 压缩训练T恤", "速干排汗，肌肉支撑，健身必选", 25900, 500, 3},
	{"运动户外专营", "迪卡侬碳纤维公路自行车", "轻量化车架，变速顺滑，骑行爱好者", 299900, 50, 3},
	{"运动户外专营", "阿迪达斯羽毛球拍套装", "专业用拍，附带羽毛球和拍袋", 19900, 300, 3},
	{"运动户外专营", "高压防水帐篷 2-3人", "铝合金骨架，速搭设计，四季通用", 59900, 150, 3},

	// 家居生活馆 (categoryId=4)
	{"家居生活馆", "乳胶枕头护颈椎", "天然乳胶，记忆支撑，改善睡眠", 29900, 600, 4},
	{"家居生活馆", "北欧风简约窗帘", "遮光隔热，提升居室格调，多色可选", 19900, 400, 4},
	{"家居生活馆", "不锈钢真空保温杯 500ml", "24 小时保温，便携防漏，多色可选", 12900, 1000, 4},
	{"家居生活馆", "香薰蜡烛礼盒套装", "天然大豆蜡，木芯燃烧，净化空气", 9900, 800, 4},
	{"家居生活馆", "竹制厨房刀架套装", "6 件套，防锈不粘，环保健康", 39900, 300, 4},
	{"家居生活馆", "升降电脑桌站立式", "电动升降，护腰防颈椎，居家办公", 199900, 100, 4},
	{"家居生活馆", "北欧风陶瓷花瓶", "简洁大方，适合各种鲜花摆放", 7900, 500, 4},

	// 美食零食铺 (categoryId=5)
	{"美食零食铺", "每日坚果混合礼盒 1kg", "无添加，每日营养，送礼自吃俱佳", 8900, 1000, 5},
	{"美食零食铺", "比利时手工巧克力礼盒", "进口可可，精致包装，节日首选", 15900, 500, 5},
	{"美食零食铺", "新疆无核葡萄干 500g", "自然晒制，甜而不腻，果肉饱满", 3900, 2000, 5},
	{"美食零食铺", "原味山姆腰果 500g", "低温烘焙，颗粒饱满，健康美味", 6900, 1500, 5},
	{"美食零食铺", "进口薯片大礼包（12袋）", "多口味组合，追剧必备，量大实惠", 5900, 800, 5},
	{"美食零食铺", "云南古树普洱茶饼 357g", "陈年存放，越陈越香，茶友首选", 29900, 300, 5},

	// 图书文具店 (categoryId=6)
	{"图书文具店", "《人类简史》尤瓦尔·赫拉利", "跨越70万年人类历史，震撼心灵", 5900, 500, 6},
	{"图书文具店", "《原则》瑞·达利欧", "对工作与生活的原则性思考", 6800, 400, 6},
	{"图书文具店", "百乐 Juice 果汁笔 20色套装", "颜色丰富，书写流畅，手账必备", 9900, 800, 6},
	{"图书文具店", "A5 点阵笔记本（子弹笔记）", "80g 道林纸，防渗墨，思维整理利器", 4900, 600, 6},
	{"图书文具店", "《深度工作》卡尔·纽波特", "如何在纷扰世界中专注工作", 5500, 450, 6},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	shopConn := sqlx.NewMysql(*shopDS)
	productConn := sqlx.NewMysql(*productDS)

	// load shop name → id mapping
	shopRows := []struct {
		Id   int64  `db:"id"`
		Name string `db:"name"`
	}{}
	if err := shopConn.QueryRowsCtx(ctx, &shopRows, "SELECT id, name FROM shop WHERE status=1"); err != nil {
		log.Fatalf("load shops: %v", err)
	}
	shopIdByName := make(map[string]int64, len(shopRows))
	for _, r := range shopRows {
		shopIdByName[r.Name] = r.Id
	}
	if len(shopIdByName) == 0 {
		log.Fatal("no shops found — run mall-shop-rpc/cmd/seed first")
	}
	fmt.Printf("[lookup] %d shops loaded\n", len(shopIdByName))

	gconn, err := grpc.NewClient(*productAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("product dial: %v", err)
	}
	defer gconn.Close()
	client := product.NewProductClient(gconn)

	created, skipped := 0, 0
	for i, p := range products {
		shopId, ok := shopIdByName[p.shopName]
		if !ok {
			log.Printf("[warn] shop %q not found, skipping %q", p.shopName, p.name)
			skipped++
			continue
		}

		var count int64
		if err := productConn.QueryRowCtx(ctx, &count, "SELECT COUNT(*) FROM product WHERE name = ?", p.name); err != nil {
			log.Fatalf("check product %q: %v", p.name, err)
		}
		if count > 0 {
			fmt.Printf("[skip] product %q already exists\n", p.name)
			skipped++
			continue
		}

		images := fmt.Sprintf("%s/product_%d.png", *minioBase, i+1)
		resp, err := client.CreateProduct(ctx, &product.CreateProductReq{
			Name:        p.name,
			Description: p.desc,
			Price:       p.price,
			Stock:       p.stock,
			CategoryId:  p.categoryId,
			Images:      images,
			ShopId:      shopId,
		})
		if err != nil {
			log.Fatalf("create product %q: %v", p.name, err)
		}
		fmt.Printf("[created] product %q id=%d shop=%d price=%d\n", p.name, resp.Id, shopId, p.price)
		created++
	}

	fmt.Printf("product seed done: created=%d skipped=%d\n", created, skipped)
}
