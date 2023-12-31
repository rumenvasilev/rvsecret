package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/log"
)

// Pluralize will take in a count and if the count is != 1 it will return the singular of the word.
func Pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// TruncateString will take an integer and cut a string at that length and append an ellipsis to it.
func TruncateString(str string, maxLength int) string {

	// match a carriage return or newline character and use that as a delimiter
	// https://regex101.com/r/gb6pcj/2
	var NewlineRegex = regexp.MustCompile(`\r?\n`)

	str = NewlineRegex.ReplaceAllString(str, " ")
	str = strings.TrimSpace(str)
	if len(str) > maxLength {
		str = fmt.Sprintf("%s...", str[0:maxLength])
	}
	return str
}

// CleanURLSpaces will take a string and replace any spaces with dashes so that is may be used in a url.
func CleanURLSpaces(dirtyStrings ...string) []string {
	var result []string
	for _, s := range dirtyStrings {
		result = append(result, strings.ReplaceAll(s, " ", "-"))
	}
	return result
}

// GenerateID will create an ID for each finding based up the SHA1 of discrete data points associated
// with the finding.
func GenerateID() string {
	return GenerateIDWithLen(10000000000)
}

func GenerateIDWithLen(i int) string {
	h := sha1.New()
	source := rand.NewSource(time.Now().UnixNano())
	randNum := rand.New(source)

	_, err := io.WriteString(h, fmt.Sprintf("%x", randNum.Intn(i)))

	if err != nil {
		fmt.Println("Not able to generate finding ID: ", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenerateSecretIDWithParams(params ...string) string {
	var str string
	for _, p := range params {
		str += p
	}

	h := sha1.New()
	_, err := io.WriteString(h, str)
	if err != nil {
		log.Log.Error("unable to generate new secret id")
		panic(fmt.Sprintf("unable to generate new secret id for string %s", str))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// get EntropyInt will calculate the entrophy based upon Shannon Entropy
func GetEntropyInt(s string) float64 {
	//Shannon Entropy calculation
	m := map[rune]float64{}
	for _, r := range s {
		m[r]++
	}
	var hm float64
	for _, c := range m {
		hm += c * math.Log2(c)
	}
	l := float64(len(s))
	res := math.Log2(l) - hm/l
	return res
}

func StringToPointer(s string) *string {
	return &s
}

func PointerToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func PointerToInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func Int64ToPointer(i int64) *int64 {
	return &i
}

// AppendIfMissing will check a slice for a value before appending it
func AppendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}

// AppendToSlice will append additional items to slice if not present already and return a new slice
// additional trim support is present, if necessary
func AppendToSlice(trim bool, input, target []string) []string {
	result := target
	for _, v := range input {
		if trim {
			v = strings.TrimSpace(v)
		}
		result = AppendIfMissing(result, v)
	}
	return result
}

func MergeMaps(source, target map[string]int) {
	if source == nil || target == nil {
		return
	}
	for k, v := range source {
		if vv, ok := target[k]; ok {
			target[k] = vv + v
		} else {
			target[k] = v
		}
	}
}
