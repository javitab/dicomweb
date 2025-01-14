package cli

import (
	"fmt"
)

func PrintHelpText() {
	// Prints available functions

	fmt.Println("\nPrinting Available CLI Arguments")

	fmt.Println("./go-web <mode> <menu>")

	// Instructions for web mode
	fmt.Println(
		"\nTo run in webserver mode:\n" +
			"     For production mode: ./go-web\n" +
			"     For debug mode: ./go-web web debug")

	// CLI Utilities

	fmt.Printf("\nTo create the superuser: ./go-web create_superuser\n" +
		"     Note: This is only available without login when no other users exist in the database" +
		"     As part of this process, groups and security points will also be migrated.")

	//Print Available modes
	for util_menu := range UtilityMenus {
		fmt.Printf("\nTo access %v utility menu: ./go-web util %v \n", util_menu, util_menu)
		fmt.Printf("   Available %v menu utilities: \n", util_menu)
		for utility := range UtilityMenus[util_menu] {
			fmt.Printf("	%v\n", utility)
		}

	}
}
