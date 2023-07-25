package core

import (
	"chatgpt-go/global"
	"chatgpt-go/initialize"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func RunServer() {

	router := initialize.Routers()

	address := global.Config.System.Address

	s := initServer(address, router)

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()
	pingServer("http://127.0.0.1" + address)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown error: %v", err)
	}

	log.Println("Server exiting")

}

func initServer(address string, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      20 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

// pingServer 确保 http server是工作的.
func pingServer(url string) error {
	for i := 0; i < 10; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			fmt.Printf("在浏览器中打开：%s\n", url)
			openWebsite(url)
			return nil
		}
		// 等待间隔
		fmt.Println(err, "Waiting for the router, retry in 1 second.")
		time.Sleep(time.Second)
	}

	return errors.New("Cannot connect to the router.")
}

// 用默认浏览器打开网站
func openWebsite(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // Linux 和其他Unix-like系统
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
