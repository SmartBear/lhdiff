package lhdiff

import (
	"fmt"
)

func ExampleLhdiff_withSmallData() {
	left := `one
two
three
four`
	right := `four
three
two
one
`

	printLhDiff(left, right)

	// Output:
	// 1 -> 4
	// 2 -> 3
	// 3 -> 2
	// 4 -> 1
}

func ExampleLhdiff_withSimilarContext() {
	left := `one
two
three
four
five foo
a b c d e
seven
eight
nine
ten
eleven
`
	right := `one
two
three
four
five
a b X Y e f
seven
eight
nine
ten
ten and a half
eleven
`

	printLhDiff(left, right)

	// Output:
	// 1 -> 1
	// 2 -> 2
	// 3 -> 3
	// 4 -> 4
	// 5 -> 5
	// 6 -> 6
	// 7 -> 7
	// 8 -> 8
	// 9 -> 9
	// 10 -> 10
	// 11 -> 12
	// 12 -> 13
}

func ExampleLhdiff_withDataFromPaper() {
	left := `public int largest (int num1, int
          num2, int num3){
  //original function
  //Function to obtain
  //largest value among numbers
     int largest = 0;

     if(num1>num2)
        largest = num1;
     else largest = num2;

     if(largest>num3)
        return largest;
     else return num3;

}
`
	right := `public int largest (int num1, int
          num2, int num3){
  //Function to obtain largest
  // value among three numbers
  //change variable names
     int value = 0;
     if(first>second)
         value = first;
     else value = second;

     if(value>third)
     {
        return value;
     }
     else return third;
}
`

	printLhDiff(left, right)

	// Output:
	// 1 -> 1
	// 2 -> 2
	// 3 -> _
	// 4 -> 3
	// 5 -> 4
	// 6 -> 6
	// 7 -> 10
	// 8 -> _
	// 9 -> _
	// 10 -> _
	// 11 -> _
	// 12 -> _
	// 13 -> 13
	// 14 -> 15
	// 15 -> _
	// 16 -> 16
	// 17 -> 17
}

func printLhDiff(left string, right string) {
	leftToRight, leftCount := Lhdiff(left, right)
	for left := 0; left < leftCount; left++ {
		if right, ok := leftToRight[left]; ok {
			fmt.Printf("%d -> %d\n", left+1, right+1)
		} else {
			fmt.Printf("%d -> _\n", left+1)
		}
	}
}
