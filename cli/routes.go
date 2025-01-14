package cli

import (
	cli_auth "github.com/javitab/go-web/cli/auth"
)

var UtilityMenus = map[string]map[string]func(){
	"auth": cli_auth.MenuFuncs,
}
