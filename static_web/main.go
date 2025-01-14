package static_web

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed *
var Web_fs embed.FS

func GetFile(name string) ([]byte, error) {
	file, err := Web_fs.ReadFile(name)
	if err != nil {
		fmt.Printf("Unable to load file %v", name)
		return nil, err
	}
	return file, err
}

func HTTPFS() (http.FileSystem, error) {
	return http.FS(Web_fs), nil
}
