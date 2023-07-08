package log

import (
	"bytes"
	"distributed/registry"
	"fmt"
	stlog "log"
	"net/http"
	"strings"
)

type clientLogger struct {
	url string
}

func SetClientLogger(ServiceUrl string, clientService registry.ServiceName) {
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	stlog.SetFlags(stlog.LstdFlags)
	stlog.SetOutput(&clientLogger{url: ServiceUrl}) // stlog.SetOutput(w io.Writer)，所以我们需要把 clientLogger 实现成一个 io.Writer
}

func (cl clientLogger) Write(data []byte) (int, error) {
	trimmedData := strings.TrimSpace(string(data)) // 我使用 TrimSpace 提前将前导、尾随空白字符，包括空行去掉
	b := bytes.NewBuffer([]byte(trimmedData))      // 如果data中有多余的空行，那么构造的bytes.Buffer对象会保留这些空行;这可能导致在后续处理中出现额外的空行
	res, err := http.Post(cl.url+"log", "text/plain", b)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message. service responded with %d - %s", res.StatusCode, res.Status)
	}
	return len(data), nil
}
