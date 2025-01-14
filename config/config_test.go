package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEmbeds(t *testing.T) {
	configLoaded := false
	_, err := GetFile("auth/groups.yaml")
	if err == nil {
		fmt.Println("Config loaded successfully")
		configLoaded = true
	}

	assert.Equal(t, configLoaded, true)

}
