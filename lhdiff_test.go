package lhdiff

func ExampleLhdiff_withUnrelatedLines() {
	left := `one
two
three
four`
	right := `un
deux
trois
quatre
`

	PrintLhdiff(left, right, 4, false)

	// Output:
	// 1,_
	// 2,_
	// 3,_
	// 4,_
}

func ExampleLhdiff_withIdenticalLines() {
	left := `one
two
three
four`

	PrintLhdiff(left, left, 4, false)

	// Output:
	// 1,1
	// 2,2
	// 3,3
	// 4,4
}

func ExampleLhdiff_withEmptyLeft() {
	right := `one
two
three
four`

	PrintLhdiff("", right, 4, false)

	// Output:
	// 1,_
}

func ExampleLhdiff_withEmptyRight() {
	left := `one
two
three
four`

	PrintLhdiff(left, "", 4, false)

	// Output:
	// 1,_
	// 2,_
	// 3,_
	// 4,_

}

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

	PrintLhdiff(left, right, 4, false)

	// Output:
	// 1,4
	// 2,3
	// 3,2
	// 4,1
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

	PrintLhdiff(left, right, 4, false)

	// Output:
	// 1,1
	// 2,2
	// 3,3
	// 4,4
	// 5,5
	// 6,6
	// 7,7
	// 8,8
	// 9,9
	// 10,10
	// 11,12
	// 12,13
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

	PrintLhdiff(left, right, 4, false)

	// Output:
	// 1,1
	// 2,2
	// 3,_
	// 4,3
	// 5,4
	// 6,6
	// 7,10
	// 8,_
	// 9,_
	// 10,_
	// 11,_
	// 12,_
	// 13,13
	// 14,15
	// 15,_
	// 16,16
	// 17,17
}

func ExampleLhdiff_withDataFromMainGo() {
	left := `package main

import (
	"flag"
	"github.com/aslakhellesoy/lhdiff"
	"io/ioutil"
)

func main() {
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	//app.Parse(os.Args[1:])
	left, _ := ioutil.ReadFile(leftFile)
	right, _ := ioutil.ReadFile(rightFile)
	lhdiff.Lhdiff(string(left), string(right))
}
`

	right := `package main

import (
	"flag"
	"github.com/aslakhellesoy/lhdiff"
	"io/ioutil"
)

func main() {
	flag.Parse()
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	left, _ := ioutil.ReadFile(leftFile)
	right, _ := ioutil.ReadFile(rightFile)
	lhdiff.PrintLhdiff(string(left), string(right))
}
`

	PrintLhdiff(left, right, 4, false)

	// Output:
	// 1,1
	// 2,2
	// 3,3
	// 4,4
	// 5,5
	// 6,6
	// 7,7
	// 8,8
	// 9,9
	// 10,11
	// 11,12
	// 12,_
	// 13,13
	// 14,14
	// 15,15
	// 16,16
	// 17,17
}

func ExampleLhdiff_withDataFromLhdiffGo() {
	left := `package lhdiff

import (
	"bytes"
	"fmt"
	"github.com/ianbruene/go-difflib/difflib"
	t "github.com/rexsimiloluwah/distance_metrics/text"
	"github.com/sourcegraph/go-diff/diff"
	"math"
	"regexp"
	"sort"
	"strings"
)

type LeftToRight map[int]int

type LineInfo struct {
	lineNumber int
	content    string
	context    string
}

//func (lineInfo LineInfo) contextSimHash() uint64 {
//	sh := simhash.NewSimhash()
//	return sh.GetSimhash(sh.NewWordFeatureSet([]byte(lineInfo.context)))
//}
//
//func (lineInfo LineInfo) contentSimHash() uint64 {
//	sh := simhash.NewSimhash()
//	return sh.GetSimhash(sh.NewWordFeatureSet([]byte(lineInfo.content)))
//}

type LinePair struct {
	left  LineInfo
	right LineInfo
}

//func (linePair LinePair) contextSimhashHammingDistance() uint8 {
//	return simhash.Compare(linePair.left.contextSimHash(), linePair.right.contextSimHash())
//}
//
//func (linePair LinePair) contentSimhashHammingDistance() uint8 {
//	return simhash.Compare(linePair.left.contentSimHash(), linePair.right.contentSimHash())
//}
//
//func (linePair LinePair) combinedSimhashSimilarity() float64 {
//	return ContextSimilarityFactor*float64(linePair.contextSimhashHammingDistance()) + ContentSimilarityFactor*float64(linePair.contentSimhashHammingDistance())
//}

func (linePair LinePair) contentNormalizedLevenshteinSimilarity() float64 {
	levensthein := t.Levensthein(linePair.left.content, linePair.right.content)
	normalizedLevenhsteinDistance := float64(levensthein) / math.Max(float64(len(linePair.left.content)), float64(len(linePair.right.content)))
	return 1 - normalizedLevenhsteinDistance
}

func (linePair LinePair) contextTfIdfCosine() float64 {
	return TfIdfCosine(linePair.left.context, linePair.right.context)
}

func (linePair LinePair) combinedSimilarity() float64 {
	return ContextSimilarityFactor*float64(linePair.contextTfIdfCosine()) + ContentSimilarityFactor*float64(linePair.contentNormalizedLevenshteinSimilarity())
}

type ByCombinedSimilarity []LinePair

func (a ByCombinedSimilarity) Len() int { return len(a) }
func (a ByCombinedSimilarity) Less(i, j int) bool {
	return a[j].combinedSimilarity() < a[i].combinedSimilarity()
}
func (a ByCombinedSimilarity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

const ContextSimilarityFactor = 0.4
const ContentSimilarityFactor = 0.6
const SimilarityThreshold = 0.43

func Lhdiff(left string, right string) (LeftToRight, []string, []string) {
	leftLines := ConvertToLinesWithoutNewLine(left)
	rightLines := ConvertToLinesWithoutNewLine(right)

	diffScript, _ := difflib.GetUnifiedDiffString(difflib.LineDiffParams{
		A:        leftLines,
		B:        rightLines,
		FromFile: "left",
		ToFile:   "right",
		Context:  3,
	})
	if diffScript != "" {
		fileDiff, _ := diff.ParseFileDiff([]byte(diffScript))

		// B. Detect Unchanged Lines
		leftToRight, leftLineNumbers, rightLineNumbers := LineNumbersFromDiff(fileDiff)

		Compute(leftLineNumbers, leftLines, rightLineNumbers, rightLines, leftToRight)
		return leftToRight, leftLines, rightLines
	} else {
		var leftToRight = make(LeftToRight)
		for leftLineNumber, _ := range leftLines {
			leftToRight[leftLineNumber] = leftLineNumber
		}
		return leftToRight, leftLines, rightLines
	}
}

func PrintLhdiff(left string, right string, lines bool) {
	leftToRight, leftLines, rightLines := Lhdiff(left, right)
	for left := 0; left < len(leftLines); left++ {
		if right, ok := leftToRight[left]; ok {
			if lines {
				fmt.Printf("%d:%s -> %d:%s\n", left+1, strings.TrimSpace(leftLines[left]), right+1, strings.TrimSpace(rightLines[right]))
			} else {
				fmt.Printf("%d -> %d\n", left+1, right+1)
			}
		} else {
			fmt.Printf("%d -> _\n", left+1)
		}
	}
}

func Compute(leftLineNumbers []int, leftLines []string, rightLineNumbers []int, rightLines []string, leftToRight LeftToRight) {
	leftLineInfos := MakeLineInfos(leftLineNumbers, leftLines, 4)
	rightLineInfos := MakeLineInfos(rightLineNumbers, rightLines, 4)

	for _, rightLineInfo := range rightLineInfos {
		pairs := make([]LinePair, len(leftLineInfos))
		for l, leftLineInfo := range leftLineInfos {
			pair := LinePair{
				left:  leftLineInfo,
				right: rightLineInfo,
			}
			pairs[l] = pair
			//fmt.Printf("%d:%s -> %d:%s\n", pair.left.lineNumber, strings.TrimSpace(pair.left.content), pair.right.lineNumber, strings.TrimSpace(pair.right.content))
			//fmt.Printf("  Distance (combined simhash) %f\n", pair.combinedSimhashSimilarity())
			//fmt.Printf("  Similarity (content levenshtein) %f\n", pair.contentNormalizedLevenshteinSimilarity())
			//fmt.Printf("  Similarity (context tf-idf cosine) %f\n", pair.contextTfIdfCosine())
			//fmt.Printf("  Distance (combined) %f\n", pair.combinedSimilarity())
			//fmt.Printf("%d:%s -> %d:%s (%f %f %f)\n",
			//	pair.left.lineNumber + 1,
			//	strings.TrimSpace(pair.left.content),
			//	pair.right.lineNumber + 1,
			//	strings.TrimSpace(pair.right.content),
			//	pair.contextTfIdfCosine(),
			//	pair.contentNormalizedLevenshteinSimilarity(),
			//	pair.combinedSimilarity(),
			//)
		}
		sort.Sort(ByCombinedSimilarity(pairs))
		pair := pairs[0]
		//fmt.Printf("%d:%s -> %d:%s (%f %f %f)\n",
		//	pair.left.lineNumber + 1,
		//	strings.TrimSpace(pair.left.content),
		//	pair.right.lineNumber + 1,
		//	strings.TrimSpace(pair.right.content),
		//	pair.contextTfIdfCosine(),
		//	pair.contentNormalizedLevenshteinSimilarity(),
		//	pair.combinedSimilarity(),
		//)
		if pair.combinedSimilarity() > SimilarityThreshold {
			leftToRight[pair.left.lineNumber] = pair.right.lineNumber
		}
	}
}

func MakeLineInfos(lineNumbers []int, lines []string, contextSize int) []LineInfo {
	lineInfos := make([]LineInfo, len(lineNumbers))
	for i, lineNumber := range lineNumbers {
		context := GetContext(lineNumber, lines, contextSize)
		content := lines[lineNumber]
		lineInfos[i] = LineInfo{
			lineNumber: lineNumber,
			context:    context,
			content:    content,
		}
	}
	return lineInfos
}

var /* const */ brackets = regexp.MustCompile("^[{()}]$")

// GetContext returns a string consisting of (up to) contextSize context lines above and below lineIndex.
// a line is considered to be a context line if it is not an "insignificant" line, i.e. either blank
// or just a curly brace or parenthesis (whitespace trimmed).
func GetContext(lineNumber int, lines []string, contextSize int) string {
	var context []string

	i := lineNumber - 1

	for j := 0; i >= 0 && j < contextSize; {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if len(trimmed) != 0 && !brackets.MatchString(trimmed) {
			context = append([]string{line}, context...)
			j++
		}
		i--
	}

	i = lineNumber + 1
	for j := 0; i < len(lines) && j < contextSize; {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if len(trimmed) != 0 && !brackets.MatchString(trimmed) {
			context = append(context, line)
			j++
		}
		i++
	}

	return strings.Join(context, "")
}

// LineNumbersFromDiff returns three collections:
// 1: a map of unchanged line numbers (from left to right)
// 2: a slice of removed line numbers in left
// 3: a slice of added line numbers in right
func LineNumbersFromDiff(fileDiff *diff.FileDiff) (LeftToRight, []int, []int) {
	var leftToRight = make(LeftToRight)

	// Deleted from left
	var leftLineNumbers []int
	// Added to right
	var rightLineNumbers []int

	for i, hunk := range fileDiff.Hunks {
		if i == 0 {
			for lineNumber := 0; lineNumber < int(hunk.OrigStartLine)-1; lineNumber++ {
				leftToRight[lineNumber] = lineNumber
			}
		}
		leftToRightHunk, leftLineNumbersHunk, rightLineNumbersHunk := LineNumbersFromHunk(hunk)
		for leftLineNumber, rightLineNumber := range leftToRightHunk {
			if existingRightLineNumber, exists := leftToRight[leftLineNumber]; exists {
				panic(fmt.Sprintf("Already mapped %d to %v. Cannot map again to %v", leftLineNumber, existingRightLineNumber, rightLineNumber))
			} else {
				leftToRight[leftLineNumber] = rightLineNumber
			}
		}
		leftLineNumbers = append(leftLineNumbers, leftLineNumbersHunk...)
		rightLineNumbers = append(rightLineNumbers, rightLineNumbersHunk...)
	}
	return leftToRight, leftLineNumbers, rightLineNumbers
}

func LineNumbersFromHunk(hunk *diff.Hunk) (LeftToRight, []int, []int) {
	var leftToRight = make(LeftToRight)
	var leftLineNumbers []int
	var rightLineNumbers []int

	lines := bytes.Split(hunk.Body, []byte{'\n'})

	var leftLineNumber = int(hunk.OrigStartLine) - 1
	var rightLineNumber = int(hunk.NewStartLine) - 1

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '-':
			leftLineNumbers = append(leftLineNumbers, leftLineNumber)
			leftLineNumber++
		case '+':
			rightLineNumbers = append(rightLineNumbers, rightLineNumber)
			rightLineNumber++
		default:
			leftToRight[leftLineNumber] = rightLineNumber
			leftLineNumber++
			rightLineNumber++
		}
	}
	return leftToRight, leftLineNumbers, rightLineNumbers
}

func ConvertToLinesWithoutNewLine(text string) []string {
	// See "A. Preprocess input files" in lhdiff-paper-long-version.pdf
	return Map(difflib.SplitLines(text), RemoveMultipleSpaceAndTrim)
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func RemoveMultipleSpaceAndTrim(s string) string {
	re := regexp.MustCompile("[ \t]+")
	return strings.TrimSpace(re.ReplaceAllString(s, " ")) + "\n"
}
`
	right := `package lhdiff

import (
	"bytes"
	"fmt"
	"github.com/ianbruene/go-difflib/difflib"
	t "github.com/rexsimiloluwah/distance_metrics/text"
	"github.com/sourcegraph/go-diff/diff"
	"math"
	"regexp"
	"sort"
	"strings"
)

type LineInfo struct {
	lineNumber int
	content    string
	context    string
}

type LinePair struct {
	left  LineInfo
	right LineInfo
}

func (linePair LinePair) contentNormalizedLevenshteinSimilarity() float64 {
	levensthein := t.Levensthein(linePair.left.content, linePair.right.content)
	normalizedLevenhsteinDistance := float64(levensthein) / math.Max(float64(len(linePair.left.content)), float64(len(linePair.right.content)))
	return 1 - normalizedLevenhsteinDistance
}

func (linePair LinePair) contextTfIdfCosineSimilarity() float64 {
	return TfIdfCosineSimilarity(linePair.left.context, linePair.right.context)
}

func (linePair LinePair) combinedSimilarity() float64 {
	return ContextSimilarityFactor*float64(linePair.contextTfIdfCosineSimilarity()) + ContentSimilarityFactor*float64(linePair.contentNormalizedLevenshteinSimilarity())
}

type ByCombinedSimilarity []*LinePair

func (a ByCombinedSimilarity) Len() int { return len(a) }
func (a ByCombinedSimilarity) Less(i, j int) bool {
	return a[j].combinedSimilarity() < a[i].combinedSimilarity()
}
func (a ByCombinedSimilarity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

const ContextSimilarityFactor = 0.4
const ContentSimilarityFactor = 0.6
const SimilarityThreshold = 0.43

func Lhdiff(left string, right string, contextSize int) []*LinePair {
	leftLines := ConvertToLinesWithoutNewLine(left)
	rightLines := ConvertToLinesWithoutNewLine(right)

	linePairs := make([]*LinePair, len(leftLines))

	diffScript, _ := difflib.GetUnifiedDiffString(difflib.LineDiffParams{
		A:        leftLines,
		B:        rightLines,
		FromFile: "left",
		ToFile:   "right",
		Context:  3,
	})
	if diffScript != "" {
		fileDiff, _ := diff.ParseFileDiff([]byte(diffScript))

		// B. Detect Unchanged Lines
		leftLineNumbers, rightLineNumbers := LineNumbersFromDiff(fileDiff, linePairs, leftLines, rightLines, contextSize)

		leftLineInfos := MakeLineInfos(leftLineNumbers, leftLines, contextSize)
		rightLineInfos := MakeLineInfos(rightLineNumbers, rightLines, contextSize)

		for _, rightLineInfo := range rightLineInfos {
			pairs := make([]*LinePair, len(leftLineInfos))
			for l, leftLineInfo := range leftLineInfos {
				pair := &LinePair{
					left:  leftLineInfo,
					right: rightLineInfo,
				}
				pairs[l] = pair
				//fmt.Printf("%d:%s -> %d:%s\n", pair.left.lineNumber, strings.TrimSpace(pair.left.content), pair.right.lineNumber, strings.TrimSpace(pair.right.content))
				//fmt.Printf("  Distance (combined simhash) %f\n", pair.combinedSimhashSimilarity())
				//fmt.Printf("  Similarity (content levenshtein) %f\n", pair.contentNormalizedLevenshteinSimilarity())
				//fmt.Printf("  Similarity (context tf-idf cosine) %f\n", pair.contextTfIdfCosineSimilarity())
				//fmt.Printf("  Distance (combined) %f\n", pair.combinedSimilarity())
				//fmt.Printf("%d:%s -> %d:%s (%f %f %f)\n",
				//	pair.left.lineNumber + 1,
				//	strings.TrimSpace(pair.left.content),
				//	pair.right.lineNumber + 1,
				//	strings.TrimSpace(pair.right.content),
				//	pair.contextTfIdfCosineSimilarity(),
				//	pair.contentNormalizedLevenshteinSimilarity(),
				//	pair.combinedSimilarity(),
				//)
			}
			sort.Sort(ByCombinedSimilarity(pairs))
			pair := pairs[0]
			//fmt.Printf("%d:%s -> %d:%s (%f %f %f)\n",
			//	pair.left.lineNumber + 1,
			//	strings.TrimSpace(pair.left.content),
			//	pair.right.lineNumber + 1,
			//	strings.TrimSpace(pair.right.content),
			//	pair.contextTfIdfCosineSimilarity(),
			//	pair.contentNormalizedLevenshteinSimilarity(),
			//	pair.combinedSimilarity(),
			//)
			if pair.combinedSimilarity() > SimilarityThreshold {
				linePairs[pair.left.lineNumber] = pair
			}
		}
	} else {
		for leftLineNumber, _ := range leftLines {
			lineInfo := MakeLineInfo(leftLineNumber, leftLines, 4)
			linePairs[leftLineNumber] = &LinePair{
				left:  lineInfo,
				right: lineInfo,
			}
		}
	}
	return linePairs
}

func PrintLhdiff(left string, right string, contextSize int, lines bool) {
	linePairs := Lhdiff(left, right, contextSize)
	for lineNumber, pair := range linePairs {
		if pair == nil {
			fmt.Printf("%d,_\n", lineNumber+1)
		} else {
			if lines {
				fmt.Printf("%d:%s,%d:%s\n", pair.left.lineNumber+1, strings.TrimSpace(pair.left.content), pair.right.lineNumber+1, strings.TrimSpace(pair.right.content))
			} else {
				fmt.Printf("%d,%d\n", pair.left.lineNumber+1, pair.right.lineNumber+1)
			}
		}
	}
}

func MakeLineInfos(lineNumbers []int, lines []string, contextSize int) []LineInfo {
	lineInfos := make([]LineInfo, len(lineNumbers))
	for i, lineNumber := range lineNumbers {
		lineInfos[i] = MakeLineInfo(lineNumber, lines, contextSize)
	}
	return lineInfos
}

func MakeLineInfo(lineNumber int, lines []string, contextSize int) LineInfo {
	content := lines[lineNumber]
	context := GetContext(lineNumber, lines, contextSize)
	lineInfo := LineInfo{
		lineNumber: lineNumber,
		context:    context,
		content:    content,
	}
	return lineInfo
}

var /* const */ brackets = regexp.MustCompile("^[{()}]$")

// GetContext returns a string consisting of (up to) contextSize context lines above and below lineIndex.
// a line is considered to be a context line if it is not an "insignificant" line, i.e. either blank
// or just a curly brace or parenthesis (whitespace trimmed).
func GetContext(lineNumber int, lines []string, contextSize int) string {
	var context []string

	i := lineNumber - 1

	for j := 0; i >= 0 && j < contextSize; {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if len(trimmed) != 0 && !brackets.MatchString(trimmed) {
			context = append([]string{line}, context...)
			j++
		}
		i--
	}

	i = lineNumber + 1
	for j := 0; i < len(lines) && j < contextSize; {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if len(trimmed) != 0 && !brackets.MatchString(trimmed) {
			context = append(context, line)
			j++
		}
		i++
	}

	return strings.Join(context, "")
}

// LineNumbersFromDiff returns three collections:
// 1: a map of unchanged line numbers (from left to right)
// 2: a slice of removed line numbers in left
// 3: a slice of added line numbers in right
func LineNumbersFromDiff(fileDiff *diff.FileDiff, pairs []*LinePair, leftLines []string, rightLines []string, contextSize int) ([]int, []int) {
	// Deleted from left
	var leftLineNumbers []int
	// Added to right
	var rightLineNumbers []int

	for i, hunk := range fileDiff.Hunks {
		if i == 0 {
			for lineNumber := 0; lineNumber < int(hunk.OrigStartLine)-1; lineNumber++ {
				lineInfo := MakeLineInfo(lineNumber, leftLines, contextSize)
				pairs[lineNumber] = &LinePair{
					left:  lineInfo,
					right: lineInfo,
				}
			}
		}
		leftLineNumbersHunk, rightLineNumbersHunk := LineNumbersFromHunk(hunk, pairs, leftLines, rightLines, contextSize)
		leftLineNumbers = append(leftLineNumbers, leftLineNumbersHunk...)
		rightLineNumbers = append(rightLineNumbers, rightLineNumbersHunk...)
	}
	return leftLineNumbers, rightLineNumbers
}

func LineNumbersFromHunk(hunk *diff.Hunk, pairs []*LinePair, leftLines []string, rightLines []string, contextSize int) ([]int, []int) {
	var leftLineNumbers []int
	var rightLineNumbers []int

	lines := bytes.Split(hunk.Body, []byte{'\n'})

	var leftLineNumber = int(hunk.OrigStartLine) - 1
	var rightLineNumber = int(hunk.NewStartLine) - 1

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '-':
			leftLineNumbers = append(leftLineNumbers, leftLineNumber)
			leftLineNumber++
		case '+':
			rightLineNumbers = append(rightLineNumbers, rightLineNumber)
			rightLineNumber++
		default:
			pairs[leftLineNumber] = &LinePair{
				left:  MakeLineInfo(leftLineNumber, leftLines, contextSize),
				right: MakeLineInfo(rightLineNumber, rightLines, contextSize),
			}
			leftLineNumber++
			rightLineNumber++
		}
	}
	return leftLineNumbers, rightLineNumbers
}

func ConvertToLinesWithoutNewLine(text string) []string {
	// See "A. Preprocess input files" in lhdiff-paper-long-version.pdf
	return Map(difflib.SplitLines(text), RemoveMultipleSpaceAndTrim)
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func RemoveMultipleSpaceAndTrim(s string) string {
	re := regexp.MustCompile("[ \t]+")
	return strings.TrimSpace(re.ReplaceAllString(s, " ")) + "\n"
}
`
	// https://github.com/aslakhellesoy/lhdiff/commit/4ae3495de0c31675940861592a3929df8154785f
	PrintLhdiff(left, right, 4, false)

	// Output:
	//1,1
	//2,2
	//3,3
	//4,4
	//5,5
	//6,6
	//7,7
	//8,8
	//9,9
	//10,10
	//11,11
	//12,12
	//13,13
	//14,14
	//15,_
	//16,_
	//17,15
	//18,16
	//19,17
	//20,18
	//21,19
	//22,20
	//23,_
	//24,_
	//25,_
	//26,_
	//27,_
	//28,_
	//29,_
	//30,_
	//31,_
	//32,_
	//33,21
	//34,22
	//35,23
	//36,24
	//37,_
	//38,_
	//39,_
	//40,_
	//41,_
	//42,_
	//43,_
	//44,_
	//45,_
	//46,_
	//47,_
	//48,_
	//49,25
	//50,26
	//51,27
	//52,_
	//53,29
	//54,30
	//55,31
	//56,32
	//57,33
	//58,34
	//59,35
	//60,36
	//61,37
	//62,38
	//63,39
	//64,40
	//65,41
	//66,42
	//67,43
	//68,_
	//69,69
	//70,70
	//71,71
	//72,72
	//73,49
	//74,50
	//75,51
	//76,52
	//77,53
	//78,54
	//79,57
	//80,58
	//81,59
	//82,_
	//83,83
	//84,84
	//85,85
	//86,86
	//87,87
	//88,66
	//89,67
	//90,68
	//91,_
	//92,70
	//93,_
	//94,_
	//95,112
	//96,_
	//97,126
	//98,115
	//99,145
	//100,_
	//101,120
	//102,122
	//103,146
	//104,124
	//105,_
	//106,_
	//107,_
	//108,130
	//109,_
	//110,132
	//111,_
	//112,_
	//113,_
	//114,_
	//115,_
	//116,134
	//117,_
	//118,_
	//119,_
	//120,114
	//121,72
	//122,73
	//123,74
	//124,75
	//125,76
	//126,77
	//127,78
	//128,79
	//129,80
	//130,_
	//131,82
	//132,83
	//133,84
	//134,85
	//135,86
	//136,87
	//137,88
	//138,89
	//139,90
	//140,91
	//141,92
	//142,93
	//143,94
	//144,95
	//145,_
	//146,146
	//147,147
	//148,148
	//149,149
	//150,101
	//151,102
	//152,103
	//153,104
	//154,105
	//155,106
	//156,107
	//157,108
	//158,133
	//159,135
	//160,136
	//161,137
	//162,_
	//163,139
	//164,140
	//165,141
	//166,149
	//167,148
	//168,150
	//169,151
	//170,152
	//171,153
	//172,154
	//173,_
	//174,155
	//175,156
	//176,157
	//177,158
	//178,_
	//179,179
	//180,180
	//181,181
	//182,182
	//183,183
	//184,184
	//185,185
	//186,186
	//187,187
	//188,188
	//189,189
	//190,190
	//191,191
	//192,192
	//193,193
	//194,194
	//195,195
	//196,196
	//197,197
	//198,198
	//199,199
	//200,200
	//201,201
	//202,202
	//203,203
	//204,204
	//205,205
	//206,206
	//207,207
	//208,208
	//209,209
	//210,210
	//211,211
	//212,193
	//213,194
	//214,195
	//215,196
	//216,_
	//217,_
	//218,197
	//219,198
	//220,199
	//221,_
	//222,222
	//223,202
	//224,203
	//225,204
	//226,_
	//227,_
	//228,_
	//229,_
	//230,_
	//231,_
	//232,_
	//233,_
	//234,_
	//235,_
	//236,_
	//237,213
	//238,214
	//239,215
	//240,216
	//241,243
	//242,218
	//243,_
	//244,_
	//245,220
	//246,221
	//247,222
	//248,_
	//249,249
	//250,250
	//251,251
	//252,252
	//253,253
	//254,254
	//255,255
	//256,256
	//257,257
	//258,258
	//259,259
	//260,260
	//261,261
	//262,237
	//263,238
	//264,239
	//265,_
	//266,244
	//267,245
	//268,246
	//269,247
	//270,248
	//271,249
	//272,250
	//273,251
	//274,_
	//275,_
	//276,_
	//277,_
	//278,_
	//279,_
	//280,_
	//281,_
	//282,_
	//283,_
	//284,_
	//285,_
	//286,_
	//287,_
	//288,_
	//289,_
	//290,_
}
