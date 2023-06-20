package lhdiff

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	levenshtein "github.com/ka-weihe/fast-levenshtein"
	"github.com/mongodb-forks/go-difflib/difflib"
	"github.com/sourcegraph/go-diff/diff"
)

type LineInfo struct {
	lineNumber int
	content    string
	context    string
}

type LinePair struct {
	left  *LineInfo
	right *LineInfo
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

type ByCombinedSimilarity []LinePair

func (a ByCombinedSimilarity) Len() int { return len(a) }
func (a ByCombinedSimilarity) Less(i, j int) bool {
	return a[j].combinedSimilarity() < a[i].combinedSimilarity()
}
func (a ByCombinedSimilarity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

const ContextSimilarityFactor = 0.4
const ContentSimilarityFactor = 0.6
const SimilarityThreshold = 0.45

/**
 * Returns a list of mappings between the lines of the left and right file.
 * Each mapping is a pair of line numbers, where -1 indicates that the line is not present in the file.
 * The mappings are sorted by the line number of the left file.
 * If includeIdenticalLines is true, then lines that are identical in both files are included in the mappings.
 * Otherwise, only lines that are not identical are included.
 * The contextSize parameter determines how many lines of context are used to determine the similarity of lines.
 * The context lines are not included in the mappings.
 * The context lines are lines that are not blank and do not consist of only curly braces or parenthesis.
 * The context lines are used to determine the similarity of lines.
 * The similarity of lines is determined by a combination of the normalized Levenshtein distance of the content of the lines and the cosine similarity of the context of the lines.
 * The similarity of lines is only considered if it is above a certain threshold.
 * The mappings are determined by first finding the unchanged lines using the difflib library.
 * Then, for each line in the right file, the most similar line in the left file is found.
 * The most similar line is the line with the highest combined similarity.
 */
func Lhdiff(left string, right string, contextSize int, includeIdenticalLines bool) ([][]uint32, error) {
	leftLines := ConvertToLinesWithoutNewLine(left)
	rightLines := ConvertToLinesWithoutNewLine(right)

	mappedRightLines := make(map[int]bool)
	allPairs := make(map[int]LinePair, 0)

	diffScript, err := difflib.GetUnifiedDiffString(difflib.LineDiffParams{
		A:        leftLines,
		B:        rightLines,
		FromFile: "left",
		ToFile:   "right",
		Context:  3,
	})
	//fmt.Println(diffScript)
	if err != nil {
		return nil, err
	}
	if diffScript != "" {
		fileDiff, err := diff.ParseFileDiff([]byte(diffScript))
		if err != nil {
			return nil, err
		}

		unchangedDiffPairs, leftLineNumbers, rightLineNumbers := LineNumbersFromDiff(fileDiff, leftLines, rightLines, contextSize)
		for _, unchangedDiffPair := range unchangedDiffPairs {
			allPairs[unchangedDiffPair.left.lineNumber] = unchangedDiffPair
			mappedRightLines[unchangedDiffPair.right.lineNumber] = true
		}

		leftLineInfos := MakeLineInfos(leftLineNumbers, leftLines, contextSize)
		rightLineInfos := MakeLineInfos(rightLineNumbers, rightLines, contextSize)

		for _, rightLineInfo := range rightLineInfos {
			var similarPairCandidates []LinePair
			for _, leftLineInfo := range leftLineInfos {
				pair := LinePair{
					left:  leftLineInfo,
					right: rightLineInfo,
				}
				similarPairCandidates = append(similarPairCandidates, pair)
			}
			sort.Sort(ByCombinedSimilarity(similarPairCandidates))
			if len(similarPairCandidates) > 0 {
				mostSimilarPair := similarPairCandidates[0]
				if mostSimilarPair.combinedSimilarity() > SimilarityThreshold {
					allPairs[mostSimilarPair.left.lineNumber] = mostSimilarPair
					mappedRightLines[mostSimilarPair.right.lineNumber] = true
				}
			}
		}
	} else {
		// The files are identical
		for leftLineNumber := range leftLines {
			lineInfo := MakeLineInfo(leftLineNumber, leftLines, 4)
			allPairs[leftLineNumber] = LinePair{
				left:  lineInfo,
				right: lineInfo,
			}
			mappedRightLines[leftLineNumber] = true
		}
	}
	rightLineNumbers := make([]int, 0)
	for rightLineNumber := range rightLines {
		_, mapped := mappedRightLines[rightLineNumber]
		if !mapped {
			rightLineNumbers = append(rightLineNumbers, rightLineNumber)
		}
	}
	return makeLineMappings(allPairs, len(leftLines), rightLineNumbers, includeIdenticalLines), nil
}

func makeLineMappings(linePairs map[int]LinePair, leftLineCount int, newRightLines []int, includeIdenticalLines bool) [][]uint32 {
	lineMappings := make([][]uint32, 0)
	for leftLineNumber := 0; leftLineNumber < leftLineCount; leftLineNumber++ {
		pair, exists := linePairs[leftLineNumber]
		if !exists {
			lineMappings = append(lineMappings, []uint32{uint32(leftLineNumber + 1), 0})
		} else {
			if includeIdenticalLines || !(pair.left.content == pair.right.content && leftLineNumber == pair.right.lineNumber) {
				lineMappings = append(lineMappings, []uint32{uint32(leftLineNumber + 1), uint32(pair.right.lineNumber + 1)})
			}
		}
	}
	for _, rightLine := range newRightLines {
		lineMappings = append(lineMappings, []uint32{0, uint32(rightLine + 1)})
	}
	return lineMappings
}

func MakeLineInfos(lineNumbers []int, lines []string, contextSize int) []*LineInfo {
	lineInfos := make([]*LineInfo, len(lineNumbers))
	for i, lineNumber := range lineNumbers {
		lineInfos[i] = MakeLineInfo(lineNumber, lines, contextSize)
	}
	return lineInfos
}

func MakeLineInfo(lineNumber int, lines []string, contextSize int) *LineInfo {
	content := lines[lineNumber]
	context := GetContext(lineNumber, lines, contextSize)
	lineInfo := &LineInfo{
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

// LineNumbersFromDiff returns two slices:
// 1: a slice of removed line numbers in left
// 2: a slice of added line numbers in right
// 3:
func LineNumbersFromDiff(fileDiff *diff.FileDiff, leftLines []string, rightLines []string, contextSize int) ([]LinePair, []int, []int) {
	var unchangedPairs []LinePair
	// Deleted from left
	var leftLineNumbers []int
	// Added to right
	var rightLineNumbers []int

	previousLeftLineNumber := 0
	previousRightLineNumber := 0
	for _, hunk := range fileDiff.Hunks {
		unchangedHunkPairs, leftLineNumbersHunk, rightLineNumbersHunk := LineNumbersFromHunk(hunk, leftLines, rightLines, previousLeftLineNumber, previousRightLineNumber, contextSize)
		leftLineNumbers = append(leftLineNumbers, leftLineNumbersHunk...)
		rightLineNumbers = append(rightLineNumbers, rightLineNumbersHunk...)
		unchangedPairs = append(unchangedPairs, unchangedHunkPairs...)
		previousLeftLineNumber = int(hunk.OrigStartLine - 1 + hunk.OrigLines)
		previousRightLineNumber = int(hunk.NewStartLine - 1 + hunk.NewLines)
	}
	// Add unchanged lines after last hunk
	leftLineNumber := previousLeftLineNumber
	rightLineNumber := previousRightLineNumber
	for leftLineNumber >= 0 && leftLineNumber < len(leftLines) {
		leftLineInfo := MakeLineInfo(leftLineNumber, leftLines, contextSize)
		rightLineInfo := MakeLineInfo(rightLineNumber, rightLines, contextSize)
		unchangedPairs = append(unchangedPairs, LinePair{
			left:  leftLineInfo,
			right: rightLineInfo,
		})
		leftLineNumber++
		rightLineNumber++
	}
	return unchangedPairs, leftLineNumbers, rightLineNumbers
}

func LineNumbersFromHunk(hunk *diff.Hunk, leftLines []string, rightLines []string, previousLeftLineNumber int, previousRightLineNumber int, contextSize int) ([]LinePair, []int, []int) {
	var unchangedPairs []LinePair
	leftLineNumbers := make([]int, 0)
	rightLineNumbers := make([]int, 0)

	leftLineNumber := previousLeftLineNumber
	rightLineNumber := previousRightLineNumber
	for leftLineNumber < int(hunk.OrigStartLine)-1 {
		leftLineInfo := MakeLineInfo(leftLineNumber, leftLines, contextSize)
		rightLineInfo := MakeLineInfo(rightLineNumber, rightLines, contextSize)
		unchangedPairs = append(unchangedPairs, LinePair{
			left:  leftLineInfo,
			right: rightLineInfo,
		})
		leftLineNumber++
		rightLineNumber++
	}

	lines := bytes.Split(hunk.Body, []byte{'\n'})

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
			unchangedPairs = append(unchangedPairs, LinePair{
				left:  MakeLineInfo(leftLineNumber, leftLines, contextSize),
				right: MakeLineInfo(rightLineNumber, rightLines, contextSize),
			})
			leftLineNumber++
			rightLineNumber++
		}
	}
	return unchangedPairs, leftLineNumbers, rightLineNumbers
}

func ConvertToLinesWithoutNewLine(text string) []string {
	if text == "" {
		return make([]string, 0)
	}
	lines := strings.SplitAfter(text, "\n")
	return Map(lines, RemoveMultipleSpaceAndTrim)
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

func PrintMappings(mappings [][]uint32) error {
	for _, mapping := range mappings {
		_, err := fmt.Printf("%s,%s\n", toString(mapping[0]), toString(mapping[1]))
		if err != nil {
			return err
		}
	}
	return nil
}

func toString(i uint32) string {
	var left string
	if i == 0 {
		left = "_"
	} else {
		left = strconv.FormatUint(uint64(i), 10)
	}
	return left
}
