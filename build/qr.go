package main

import (
	"bytes"
	"fmt"
	"github.com/kbinani/screenshot"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	// 注册一个路由器
	http.HandleFunc("/qr", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.Copy(w, qr2())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		path := "/home/app/data/" + header.Filename
		out, err := os.Create(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		filePath := "/home/app/WeChat Files/" + path
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "File not found.", http.StatusNotFound)
			return
		}
		defer file.Close()
		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
			return
		}
		// 设置 Content-Type 和 Content-Length
		contentType := http.DetectContentType(make([]byte, 0, 512))
		contentLength := fmt.Sprintf("%d", fileInfo.Size())
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", contentLength)
		// 启动未缓存的响应
		http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
	})
	// 启动一个 http 服务，监听本地 5556 端口
	http.ListenAndServe(":5556", nil)
}
func qr2() *bytes.Buffer {
	//刷新
	exec.Command("sh", "-c", "wxWid=$(xdotool search '微信') && xdotool windowmove $wxWid 0 0 && xdotool mousemove 135 333 && xdotool click 1").Run()
	img, err := screenshot.Capture(65, 103, 153, 153)
	if err != nil {
		panic(err)
	}
	//获取所有活动屏幕
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		panic("没有发现活动的显示器")
	}
	buffer := bytes.NewBuffer([]byte{})
	// 将图片编码为 JPEG 并保存到文件
	if err := jpeg.Encode(buffer, img, nil); err != nil {
		panic(err)
	}
	return buffer
}
