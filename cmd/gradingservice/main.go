package main

import (
	"context"
	"distributed/grades"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddress := fmt.Sprintf("http://%v:%v", host, port) // PostUrl
	r := registry.Registration{
		ServiceName:      registry.GradingService, // 这个 ServiceName 必须和 服务发现 的保持一致
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateUrl: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}
	ctx, err := service.Start(
		context.Background(), // Background returns a non-nil, empty Context. It is never canceled, has no values, and has no deadline.
		host,
		port,
		/*
			log.RegisterHandlers()被 Start() 视为 函数本身 作为 值 传入，而不是函数调用，所以不能加"()"
				- 注意：其实不完全是函数本身，而是被复制为一个新的 函数变量，
				- 所以对它的动作，不会修改原始函数。
			调用 Start 函数时，所有的参数都会在函数调用瞬间被传递给 Start 函数，包括 RegisterHandlers 参数。
				- RegisterHandlers 这个函数只有在后续代码中 调用的时候 被执行
				- 如果 RegisterHandlers 这个函数在后续没被调用，则它不会产生任何影响
		*/
		r,
		/*	！！！
			报错Error retrieving students:  json: cannot unmarshal number into Go value of type grades.Students
			我还一直以为是Grade包哪里写错了造成数据无法正常序列化
			甚至去portal打桩测试 想尝试在哪里defer出去的
			↓ 结果是这里之前注册成其他服务了，我踏马居然没注册GradingService ↓
		*/
		grades.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatalln(err) // Fatalln等价于{l.Println(v...); os.Exit(1)}
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		// 如果没发生错误，输出以下；然后设置 logger
		fmt.Printf("Logging service found at: %s\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}
	/*
		<-ctx.Done 阻塞
		何时Done？
			1.第一个 Goroutine 发现启动 http服务器 失败，cancel()
			2.第二个 Goroutine 捕捉到了用户按下任意键，cancel()
	*/
	<-ctx.Done()
	fmt.Println("Shutting down log service.")
	/*
		可以写成以下方式，让 select 在多个通道操作中选择一个已经就绪的操作执行
		select {
		case <-ctx.Done(): // 等待context的信号:
			fmt.Println("Shutting down log service.")
		}
	*/

}
