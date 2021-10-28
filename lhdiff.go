package lhdiff

import (
	"bytes"
	"fmt"
	"github.com/ianbruene/go-difflib/difflib"
	t "github.com/rexsimiloluwah/distance_metrics/text"
	"github.com/sourcegraph/go-diff/diff"
	"math"
	"regexp"
	"sort"

	//"sort"
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
//func (linePair LinePair) combinedSimhashDistance() float64 {
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

func (a ByCombinedSimilarity) Len() int      { return len(a) }
func (a ByCombinedSimilarity) Less(i, j int) bool {
	return a[j].combinedSimilarity() < a[i].combinedSimilarity()
}
func (a ByCombinedSimilarity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

const ContextSimilarityFactor = 0.4
const ContentSimilarityFactor = 0.6
const SimilarityThreshold = 0.40

func Lhdiff(left string, right string) (LeftToRight, int) {
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
		return leftToRight, len(leftLines)
	} else {
		var leftToRight = make(LeftToRight)
		for leftLineNumber, _ := range leftLines {
			leftToRight[leftLineNumber] = leftLineNumber
		}
		return leftToRight, len(leftLines)
	}
}

func PrintLhdiff(left string, right string) {
	leftToRight, leftCount := Lhdiff(left, right)
	for left := 0; left < leftCount; left++ {
		if right, ok := leftToRight[left]; ok {
			fmt.Printf("%d -> %d\n", left+1, right+1)
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
			//fmt.Printf("  Distance (combined simhash) %f\n", pair.combinedSimhashDistance())
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
		//	pair.left.lineNumber,
		//	strings.TrimSpace(pair.left.content),
		//	pair.right.lineNumber,
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
			for lineNumber := 0; lineNumber < int(hunk.OrigStartLine) - 1; lineNumber++ {
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
