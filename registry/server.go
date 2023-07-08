// 注册服务的 WEB service
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

const ServiceHost = "localhost"
const ServicePort = "3000"
const ServicesUrl = "http://" + ServiceHost + ":" + ServicePort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex // 保证在并发访问的时候，Registration 是线程安全的
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
} // 建立一个包级reg变量

/* 用于注册 */
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	// 在服务注册时刚好，把需要的依赖服务请求过来
	err := r.SendRequiredServices(reg)
	return err
}

// 把需要的依赖服务请求过来
func (r registry) SendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	// 找到 某个服务所有 RequiredServices （依赖），把它们全部添加到Added中
	for _, service := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if service.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: service.ServiceName,
					Url:  service.ServiceUrl,
				})
			}
		}
	}

	err := r.sendPatch(p, reg.ServiceUpdateUrl)
	if err != nil {
		return err
	}
	return nil
}

func (r registry) sendPatch(p patch, url string) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	// NewBuffer使用buf作为初始内容创建并初始化一个Buffer。本函数用于创建一个用于读取已存在数据的buffer；
	// 也用于指定用于写入的内部缓冲的大小，此时，buf应为一个具有指定容量但长度为0的切片。buf会被作为返回值的底层缓冲切片。
	_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	return nil
}

/* 移除url对应的服务 */
func (r *registry) remove(url string) error {
	// 去注册表找有没有 指定url，有就去掉
	for i := range reg.registrations {
		if reg.registrations[i].ServiceUrl == url {
			// ！！！千万记得上锁
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...) // 去掉i的内容，相当于 => i前面的内容+i后面的内容
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

/*
	创建 WEB Service

*/
type RegistrationService struct{}

func (s RegistrationService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received.")
	switch r.Method { // 注册服务使用POST
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r) // 解码
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// 如果没出错，打印 服务 和 URL
		log.Printf("Adding service: %v with %v", r.ServiceName, r.ServiceUrl)
		err = reg.add(r) // 注册信息添到 reg.registrations 里（add方法里有互斥锁，防止并发时死锁）
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete: // 移除服务使用DELETE
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError) //处理请求时发生了错误
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL : %s", url)
		err = reg.remove(url) // 移除url对应的服务
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError) // remove时发生了错误
			return
		}

	default:
		// 因为我只写了 POST 方法，所以其他方法禁止
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
