package config

import (
	"embed"
	"fmt"
)

//go:embed *
var config embed.FS

func GetFile(name string) ([]byte, error) {
	file, err := config.ReadFile(name)
	if err != nil {
		fmt.Printf("Unable to load file %v", name)
		return nil, err
	}
	return file, err
}
