package lhdiff

import (
	"fmt"
)

func ExampleGetContext_withoutSpecialCharacters() {
	lines := []string{
		"0\n",
		"1\n",
		"2\n",
		"3\n",
		"4\n",
		"5\n",
		"6\n",
		"7\n",
		"8\n",
		"9\n",
	}

	context := GetContext(5, lines, 3)
	fmt.Println(context)

	// Output:
	// 2
	// 3
	// 4
	// 6
	// 7
	// 8
}

func ExampleGetContext_withSpacesAndBrackets() {
	lines := []string{
		"0\n",
		"1\n",
		"{\n",
		")\n",
		"4\n",
		"5\n",
		"\n",
		"  \n",
		"8\n",
		"9\n",
	}

	context := GetContext(5, lines, 3)
	fmt.Println(context)

	// Output:
	// 0
	// 1
	// 4
	// 8
	// 9
}
