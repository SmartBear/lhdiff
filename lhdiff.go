package lhdiff

import (
	"bytes"
	"fmt"
	"github.com/go-dedup/simhash"
	"github.com/ianbruene/go-difflib/difflib"
	"github.com/ka-weihe/fast-levenshtein"
	"github.com/sourcegraph/go-diff/diff"
	"math"
	"regexp"
	"sort"
	"strings"
)

type LineInfo struct {
	lineNumber int32
	content    string
	context    string
}

func (lineInfo LineInfo) contextSimHash() uint64 {
	sh := simhash.NewSimhash()
	return sh.GetSimhash(sh.NewWordFeatureSet([]byte(lineInfo.context)))
}

func (lineInfo LineInfo) contentSimHash() uint64 {
	sh := simhash.NewSimhash()
	return sh.GetSimhash(sh.NewWordFeatureSet([]byte(lineInfo.content)))
}

type LinePair struct {
	left  LineInfo
	right LineInfo
}

func (linePair LinePair) contextSimhashHammingDistance() uint8 {
	return simhash.Compare(linePair.left.contextSimHash(), linePair.right.contextSimHash())
}

func (linePair LinePair) contentSimhashHammingDistance() uint8 {
	return simhash.Compare(linePair.left.contentSimHash(), linePair.right.contentSimHash())
}

func (linePair LinePair) combinedSimhashSimilarity() float64 {
	return ContextSimilarityFactor*float64(linePair.contextSimhashHammingDistance())/HammingFactor + ContentSimilarityFactor*float64(linePair.contentSimhashHammingDistance())/HammingFactor
}

func (linePair LinePair) contentNormalizedLevenshteinSimilarity() float64 {
	distance := levenshtein.Distance(linePair.left.content, linePair.right.content)
	normalizedLevenhsteinDistance := float64(distance) / math.Max(float64(len(linePair.left.content)), float64(len(linePair.right.content)))
	return 1 - normalizedLevenhsteinDistance
}

func (linePair LinePair) contextTfIdfCosineSimilarity() float64 {
	return TfIdfCosineSimilarity(linePair.left.context, linePair.right.context)
}

func (linePair LinePair) combinedSimilarity() float64 {
	contentSimilarity := linePair.contentNormalizedLevenshteinSimilarity()
	if contentSimilarity <= 0.5 {
		return 0.0
	}
	contextSimilarity := linePair.contextTfIdfCosineSimilarity()
	return ContentSimilarityFactor*contentSimilarity + ContextSimilarityFactor*contextSimilarity
}

type BySimhashSimilarity []*LinePair

func (a BySimhashSimilarity) Len() int { return len(a) }
func (a BySimhashSimilarity) Less(i, j int) bool {
	return a[j].combinedSimhashSimilarity() < a[i].combinedSimhashSimilarity()
}
func (a BySimhashSimilarity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ByCombinedSimilarity []*LinePair

func (a ByCombinedSimilarity) Len() int { return len(a) }
func (a ByCombinedSimilarity) Less(i, j int) bool {
	return a[j].combinedSimilarity() < a[i].combinedSimilarity()
}
func (a ByCombinedSimilarity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

const ContextSimilarityFactor = 0.4
const ContentSimilarityFactor = 0.6
const SimilarityThreshold = 0.43
const HammingFactor = 32.0
const CandidateSize = 15

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
	// fmt.Println(diffScript)
	if diffScript != "" {
		fileDiff, _ := diff.ParseFileDiff([]byte(diffScript))

		// B. Detect Unchanged Lines
		leftLineNumbers, rightLineNumbers := LineNumbersFromDiff(fileDiff, linePairs, leftLines, rightLines, contextSize)

		leftLineInfos := MakeLineInfos(leftLineNumbers, leftLines, contextSize)
		rightLineInfos := MakeLineInfos(rightLineNumbers, rightLines, contextSize)

		for _, rightLineInfo := range rightLineInfos {
			var pairs []*LinePair
			//pairs := make([]*LinePair)
			for _, leftLineInfo := range leftLineInfos {
				pair := &LinePair{
					left:  leftLineInfo,
					right: rightLineInfo,
				}
				//distance := pair.combinedSimhashSimilarity()
				//fmt.Printf("Simash distance: %f\n", distance)
				//pairs[l] = pair
				pairs = append(pairs, pair)
			}
			sort.Sort(ByCombinedSimilarity(pairs))
			if len(pairs) > 0 {
				pair := pairs[0]
				if pair.combinedSimilarity() > SimilarityThreshold {
					linePairs[pair.left.lineNumber] = pair
				}
			}
		}
	} else {
		for leftLineNumber, _ := range leftLines {
			lineInfo := MakeLineInfo(int32(leftLineNumber), leftLines, 4)
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

func MakeLineInfos(lineNumbers []int32, lines []string, contextSize int) []LineInfo {
	lineInfos := make([]LineInfo, len(lineNumbers))
	for i, lineNumber := range lineNumbers {
		lineInfos[i] = MakeLineInfo(int32(lineNumber), lines, contextSize)
	}
	return lineInfos
}

func MakeLineInfo(lineNumber int32, lines []string, contextSize int) LineInfo {
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
func GetContext(lineNumber int32, lines []string, contextSize int) string {
	var context []string

	i := int(lineNumber) - 1

	for j := 0; i >= 0 && j < contextSize; {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if len(trimmed) != 0 && !brackets.MatchString(trimmed) {
			context = append([]string{line}, context...)
			j++
		}
		i--
	}

	i = int(lineNumber) + 1
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
func LineNumbersFromDiff(fileDiff *diff.FileDiff, pairs []*LinePair, leftLines []string, rightLines []string, contextSize int) ([]int32, []int32) {
	// Deleted from left
	var leftLineNumbers []int32
	// Added to right
	var rightLineNumbers []int32

	previousLeftLineNumber := int32(0)
	previousRightLineNumber := int32(0)
	for _, hunk := range fileDiff.Hunks {
		leftLineNumbersHunk, rightLineNumbersHunk := LineNumbersFromHunk(hunk, pairs, leftLines, rightLines, previousLeftLineNumber, previousRightLineNumber, contextSize)
		leftLineNumbers = append(leftLineNumbers, leftLineNumbersHunk...)
		rightLineNumbers = append(rightLineNumbers, rightLineNumbersHunk...)
		previousLeftLineNumber = hunk.OrigStartLine - 1 + hunk.OrigLines
		previousRightLineNumber = hunk.NewStartLine - 1 + hunk.NewLines
	}
	// Add unchanged lines after last hunk
	leftLineNumber := previousLeftLineNumber
	rightLineNumber := previousRightLineNumber
	for int(leftLineNumber) < len(leftLines) {
		leftLineInfo := MakeLineInfo(int32(leftLineNumber), leftLines, contextSize)
		rightLineInfo := MakeLineInfo(rightLineNumber, rightLines, contextSize)
		pairs[leftLineNumber] = &LinePair{
			left:  leftLineInfo,
			right: rightLineInfo,
		}
		leftLineNumber++
		rightLineNumber++
	}
	return leftLineNumbers, rightLineNumbers
}

func LineNumbersFromHunk(hunk *diff.Hunk, pairs []*LinePair, leftLines []string, rightLines []string, previousLeftLineNumber int32, previousRightLineNumber int32, contextSize int) ([]int32, []int32) {
	var leftLineNumbers []int32
	var rightLineNumbers []int32

	leftLineNumber := previousLeftLineNumber
	rightLineNumber := previousRightLineNumber
	for leftLineNumber < hunk.OrigStartLine-1 {
		leftLineInfo := MakeLineInfo(leftLineNumber, leftLines, contextSize)
		rightLineInfo := MakeLineInfo(rightLineNumber, rightLines, contextSize)
		pairs[leftLineNumber] = &LinePair{
			left:  leftLineInfo,
			right: rightLineInfo,
		}
		leftLineNumber++
		rightLineNumber++
	}

	lines := bytes.Split(hunk.Body, []byte{'\n'})

	//var leftLineNumber = int(hunk.OrigStartLine) - 1
	//var rightLineNumber = int(hunk.NewStartLine) - 1

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
