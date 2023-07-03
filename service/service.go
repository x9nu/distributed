/*
	服务集中化管理
*/

package service

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

/* 启动服务 */
func Start(ctx context.Context, host, port string, reg registry.Registration, registerHandlerFunc func()) (context.Context, error) {
	// 注册处理器
	registerHandlerFunc()
	ctx = startService(ctx, reg.ServiceName, host, port) // 启动 Web service

	/* 调用 POST 以注册服务 */
	err := registry.RegisterService(reg)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// service.ServiceName 字段。这种写法不是语法糖，而是利用结构体的字段选择器来直接访问结构体中的字段
func startService(ctx context.Context, serviceName registry.ServiceName, host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = ":" + port // 本地+端口号

	go func() {
		log.Println(srv.ListenAndServe()) // 启动时出现错误，打印
		/* 关闭服务 */
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		cancel()
	}() // IIFE 立即调用表达式，这是一种设计模式（Invoking the anonymous function immediately.the "()"" at the end of anonymous function）

	go func() {
		fmt.Printf("%v started. Press any key to stop. \n", serviceName)
		/* var s string 和 fmt.Scanln(&s) 这两句话表示，如果接收到了任何按键，代码就会接着向下走*/
		var s string
		fmt.Scanln(&s)
		/* 关闭服务 */
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}

		srv.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
