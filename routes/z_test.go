package routes

import (
	"fmt"
	"runtime"
	"testing"
)

func TestXxx(t *testing.T) {
	chatStorage, err := NewChatStorage(`D:\GOPATH\src\chatgpt-go\database.sqlite`)
	if err != nil {
		panic(err)
	}
	chatStorage.GetContextMessages("52ffa8ad-3a30-41f1-8a85-fa8aaf8ccc3d")
}
func TestArch(t *testing.T) {
	fmt.Println(runtime.GOOS)
	fmt.Println(runtime.GOARCH)

}
