package proc

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

import (
	"github.com/ipfs-force-community/gosf/logger"
)

var uptime time.Time

func init() {
	uptime = time.Now()
}

// 编译器注入的信息
var (
	App     string = unknown
	Version string = unknown
	Commit  string = unknown
)

// AppName executable name
func AppName() string {
	if App != unknown {
		return App
	}

	if len(os.Args) > 0 {
		return filepath.Base(os.Args[0])
	}

	return unknown
}

// Print process info logging
func Print() {
	logger.LS().Infof("process info, app=%s, version=%s, commit=%s, host=%s", AppName(), Version, Commit, Hostname())
}

// ServeVersion 展示服务版本
func ServeVersion(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(fmt.Sprintf("App: %s\nVersion: %s\nCommit: %s\nUptime: %s\n", AppName(), Version, Commit, time.Since(uptime))))
}

// RegisterVersionHandler 注册 version handler
func RegisterVersionHandler(mux *http.ServeMux) {
	mux.Handle("/_version", http.HandlerFunc(ServeVersion))
}
