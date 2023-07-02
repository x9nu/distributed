/* 日志服务独立运行 */
package main

import (
	"context"
	"distributed/log"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	log.Run("./ditributed.log")
	host, port := "localhost", "4000"
	ctx, err := service.Start(
		context.Background(), // Background returns a non-nil, empty Context. It is never canceled, has no values, and has no deadline.
		"Log Service",
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
		log.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatalln(err) // Fatalln等价于{l.Println(v...); os.Exit(1)}
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
