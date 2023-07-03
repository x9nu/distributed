/*
	服务注册
	比如，logservice。调用这里的函数进行自动注册
	步骤：
		- 先启动 registryservice
		- logservice 启动时， 集中服务管理Start() 会调用 client的RegisterService()
		- RegisterService()被Handle接受，ServeHTTP将请求派遣到与请求的URL最匹配的模式对应的处理器Handle。

			URL 是 http://localhost:3000/services 请求如下
			{
    			"ServiceName":"Log Service",
    			"ServiceUrl":"http://localhost:4000/log"
			}

		- 最后成功效果 => Adding service: Log Service with http://localhost:4000
*/
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func RegisterService(r Registration) error {
	buf := new(bytes.Buffer) // 开辟一个buf 实现 io.Reader
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil { // 编码发生错误
		return err
	}

	res, err := http.Post(ServicesUrl, "application/json", buf)
	if err != nil { // post 请求错误
		return err
	}

	if res.StatusCode != http.StatusOK { // 状态码不是200，仍然有错
		return fmt.Errorf("failed to register service. registry service responsed with code %v", res.StatusCode)
	}

	// 如果上述三种错误都没发生
	return nil
}
