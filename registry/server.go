// 注册服务的 WEB service
package registry

import (
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
	mutex         *sync.Mutex // 保证在并发访问的时候，Registration 是线程安全的
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.Mutex),
} // 建立一个包级reg变量

/* 用于注册 */
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
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
