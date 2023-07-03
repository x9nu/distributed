/*
	服务注册
	比如，logservice。调用这里的函数进行自动注册
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
