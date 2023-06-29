/*
	日志服务的后端逻辑
	- 该日志服务是 web 服务
	- 接收进来的 post 请求，然后把 post 请求的内容写入到 log 里
*/

package log

import (
	"io/ioutil"
	stlog "log" // 由于标准库里也有名为 log 的包，所以把标准库的取个别名
	"net/http"
	"os"
)

var log *stlog.Logger

type filelog string // filelog的目的，就是把日志写进文件系统（因为内置类型不能直接绑定方法，所以新写一个类型）

// 该方法 将数据写入到文件里
func (fl filelog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_RDONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(data)
}

/*
	创建一个自定义的 logger ，它会把日志写入到 destination 地址
	前缀是go，包含日期和时间（LstdFlags）
*/
func Run(destination string) {
	log = stlog.New(filelog(destination), "go", stlog.LstdFlags) // LstdFlags = Ldate | Ltime
}

func RegisterHandlers() {
	// 针对 "/log" 的路径做处理
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// 只处理 POST 请求，把 POST 请求中 Body 的内容读出来，通过 write 函数写入到log里
		case http.MethodPost:
			msg, err := ioutil.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		// 不是 POST 请求就返回 405 不允许的请求
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func write(message string) {
	log.Printf("%v\n", message)
}
