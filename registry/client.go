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
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

type providers struct {
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

var prov = providers{
	services: make(map[ServiceName][]string), // 键是ServiceName，值是string切片
	mutex:    new(sync.RWMutex),
}

func (prv *providers) Update(pat patch) {
	prv.mutex.Lock()
	defer prv.mutex.Unlock()
	// 两种情况
	// 1.新增
	for _, patchEntry := range pat.Added {
		if _, ok := prv.services[patchEntry.Name]; !ok { // 这种写法是Go的条件表达式写法之一，
			/*
				if statement; condition {
				}
				表达式 prv.services[patchEntry.Name] 返回了两个值：1.映射中键为 patchEntry.Name 的值	2.一个布尔值，表示该键是否存在
			*/
			prv.services[patchEntry.Name] = make([]string, 0)
		}
		// 将 patchEntry.Url 添加进 键为prv.services[patchEntry.Name] 的切片中
		prv.services[patchEntry.Name] = append(prv.services[patchEntry.Name], patchEntry.Url)
	}
	// 2.移除
	// 在patch的removed里找，如果providerUrl存在，则将他们对比，相同的去掉
	for _, patchEntry := range pat.Removed {
		if providerUrls, ok := prv.services[patchEntry.Name]; ok {
			for i := range providerUrls {
				if providerUrls[i] == patchEntry.Url {
					prv.services[patchEntry.Name] = append(providerUrls[:i], providerUrls[i+1:]...)
				}
			}
		}
	}
}

func (prv providers) get(name ServiceName) (string, error) {
	providersValue, ok := prv.services[name]
	if !ok {
		return "", fmt.Errorf("no providers available for service %v", name)
	}
	index := int(rand.Float32() * float32(len(providersValue)))
	return providersValue[index], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

type serviceUpdateHandler struct{}

func (suh serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	/*	解析json数据
		1.ioutil.ReadAll和json.Unmarshal的组合方式
		2.json.Decode方法
	*/
	var p patch
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	prov.Update(p)
}

func RegisterService(r Registration) error {
	/* 更新动作 */
	serviceUpdateUrl, err := url.Parse(r.ServiceUpdateUrl) // 解析成 url 类型
	if err != nil {
		return err
	}
	http.Handle(serviceUpdateUrl.Path, &serviceUpdateHandler{})

	/* 以下是注册 */
	buf := new(bytes.Buffer) // 开辟一个buf 实现 io.Reader
	enc := json.NewEncoder(buf)
	err = enc.Encode(r)
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

/* 关闭服务
应该放在 Service包 的 startService() 两个取消的协程中 */
func ShutdownService(url string) error {
	req, err := http.NewRequest(http.MethodDelete, ServicesUrl, bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister service. Registry service responsed with code %v", res.StatusCode)
	}
	return nil // 如果都没错就取消成功
}
