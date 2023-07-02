// 注册服务的 WEB service
package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const ServiceHost = "localhsot"
const ServicePort = "3000"
const ServicesUrl = "http://" + ServiceHost + ":" + ServicePort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.Mutex // 保证在并发访问的时候，Registration 是线程安全的
}

/* 用于注册 */
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	return nil
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.Mutex),
} // 建立一个包级reg变量

/*
	创建 WEB Service

*/
type RegistrationService struct{}

func (s RegistrationService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received.")
	switch r.Method {
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
	default:
		// 因为我只写了 POST 方法，所以其他方法禁止
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
