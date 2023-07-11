/* 服务注册独立运行 */
package main

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

func main() {
	registry.SetupRegistryService()
	http.Handle("/services", registry.RegistrationService{})

	ctx, cancel := context.WithCancel(context.Background()) // WithCancel()第二个return 是一个函数：func() { c.cancel(true, Canceled) }
	defer cancel()
	var srv http.Server
	/*
		为什么这里的地址只填端口？
		- 接下来使用的 ListenAndServe()方法 里填的确实是端口
		- ListenAndServe()方法里 调用了net.Listen，
			- Listen函数有一部分注解：一个未指定的文字IP地址，侦听所有可用的，本地系统的单播和任意播IP地址。
	*/
	srv.Addr = ":" + registry.ServicePort

	go func() {
		log.Println(srv.ListenAndServe()) // 启动出错打印
		cancel()
	}()

	go func() {
		fmt.Println("Registry started. Press any key to stop.")
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done() // 阻塞等待两个 Goroutine 的信号（有一个即可）
	fmt.Println("Shutting down registry service.")
}
