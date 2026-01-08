package main

import (
	"flag"
	"fmt"

	"arbitragex/restful/engine/internal/config"
	"arbitragex/restful/engine/internal/handler"
	"arbitragex/restful/engine/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/engine-api.yaml", "the config file")

// main 套利引擎服务的主入口函数
// 职责：
//  1. 加载配置文件
//  2. 初始化 REST 服务器
//  3. 注册路由（使用 goctl 生成的 RegisterHandlers）
//  4. 启动套利引擎服务
func main() {
	flag.Parse()

	// 加载配置文件
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 创建 REST 服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册 goctl 生成的路由
	handler.RegisterHandlers(server, ctx)

	// 启动服务器
	fmt.Printf("Starting arbitrage engine service at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
