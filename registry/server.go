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
	"time"
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
	// 在服务注册时刚好，把需要的依赖服务请求过来,比如在启动Grading时，将Grading的依赖拉过来
	err := r.SendRequiredServices(reg)

	// 服务在注册时，可以通知其他依赖它的服务（可以仔细想想这一段和上一段注释对应的代码先后顺序）
	r.notify(patch{ // 把要新增的服务作为参数
		Added: []patchEntry{
			/* patchEntry{ // redundant type from array, slice, or map composite literal 翻译：数组、切片或映射复合文本中的冗余类型，值不用再声明类型
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			}, */
			// 像这样直接 {}，取代 => patchEntry{} （类型{}）
			{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})

	return err
}

func (r registry) notify(fullPatch patch) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	/*
		- 在每个服务里遍历
			- 给每个服务开一个 Goroutine
				- 在每个服务的 RequiredServices 里遍历
					- 初始化新增/移除 p列表（patch）
					- 初始化 sendUpdate = false
					- 在 fullPatch 里遍历新增/移除的服务名
						- 如果 这个新增/移除的服务名 == RequiredServices
							- 将 需要增加/移除的 Patch 添加到 p列表的增加/移除列表
							- 赋值 sendUpdate = true
					- 如果 sendUpdate == true
						- 使用 sendPatch()
	*/
	for _, reg := range r.registrations {
		go func(reg Registration) {
			for _, reqService := range reg.RequiredServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateUrl)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
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
					URL:  service.ServiceURL,
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

/*  */
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
		if reg.registrations[i].ServiceURL == url {
			r.notify(patch{ // 把要被移除的服务作为参数
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})
			// ！！！千万记得上锁
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...) // 去掉i的内容，相当于 => i前面的内容+i后面的内容
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

// 定时检查心跳 => 使用 Goroutine 检查服务是否存在
func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)                   // 添加等待计数值
			go func(reg Registration) { // 每个服务开一个 Goroutine
				defer wg.Done() // 在线程结束时，Done方法减少WaitGroup计数器的值
				success := true
				for attempts := 0; attempts < 3; attempts++ { // 给三次尝试[检查]
					res, err := http.Get(reg.HeartbeatURL)
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						log.Printf("[Heartbeat-check] PASSED for %v", reg.ServiceName)
						// 通过检查，执行，break
						if !success { // 如果上次检查失败 => 那么success==false，也就是被卸载过，现在检查通过可以add了
							r.add(reg)
						}
						success = true
						break
					}
					if success {
						log.Printf("[Heartbeat-check] FAILED for %v", reg.ServiceName)
						// 发现有错，卸载服务
						success = false
						r.remove(reg.ServiceURL)
					}
					time.Sleep(1 * time.Second)
				}
			}(reg)
			wg.Wait() // 阻塞 直到WaitGroup计数器减为0
			time.Sleep(freq)
		}
	}
}

// 使用 once ,让程序开始时启动一次 heartbeat 方法[它是 for 无终止循环]
var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
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
		log.Printf("Adding service: %v with %v", r.ServiceName, r.ServiceURL)
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
