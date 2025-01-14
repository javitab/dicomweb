package cli

import (
	"fmt"
	"os"
)

func ExecUtilMenu(modes map[string]func()) {
	// ### ###
	// ### ### Print Menu
	// ### ###
	var modeSelected string
	var modeSelectFuncMap = map[string]func(){}

	fmt.Println("\nSelect Utility to Run:")
	var iter int
	for key := range modes {
		iter++
		fmt.Printf("# %v: %v\n", iter, key)
		modeSelectFuncMap[fmt.Sprintf("%v", iter)] = modes[key]
	}

	// ### ###
	// ### ### Add Additional Options
	// ### ###
	var addlModes = map[string]func(){
		"Exit": nil,
	}
	for key := range addlModes {
		iter++
		fmt.Printf("# %v: %v\n", iter, key)
		modeSelectFuncMap[fmt.Sprintf("%v", iter)] = modes[key]
	}

	// ### ###
	// ### ### Execute Utility
	// ### ###
	fmt.Print("Enter utility #: ")
	fmt.Scanln(&modeSelected)
	ExecFunc := modeSelectFuncMap[modeSelected]
	if ExecFunc == nil {
		os.Exit(0)
	}

	if ExecFunc != nil {
		ExecFunc()
		var cont string
		fmt.Printf("\nContinue? (y/n): ")
		fmt.Scanln(&cont)
		for cont == "y" {
			ExecFunc()
			fmt.Printf("\nContinue? (y/n): ")
			fmt.Scanln(&cont)
		}
		// helpers.ClearScreen()
		ExecUtilMenu(modes)

	}
	os.Exit(0)

}
