/*
	服务集中化管理
*/

package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

/* 启动服务 */
func Start(ctx context.Context, serviceName, host, port string, registerHandlerFunc func()) (context.Context, error) {
	// 注册处理器
	registerHandlerFunc()
	ctx = startService(ctx, serviceName, host, port) // 启动service

	return ctx, nil
}

func startService(ctx context.Context, serviceName, hsot, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = ":" + port // 本地+端口号

	go func() {
		log.Println(srv.ListenAndServe()) // 启动时出现错误，打印
		cancel()
	}() // IIFE 立即调用表达式，这是一种设计模式（Invoking the anonymous function immediately.the "()"" at the end of anonymous function）

	go func() {
		fmt.Printf("%v started.Press any key to stop. \n", serviceName)
		/* var s string 和 fmt.Scanln(&s) 这两句话表示，如果接收到了任何按键，代码就会接着向下走*/
		var s string
		fmt.Scanln(&s)

		srv.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
