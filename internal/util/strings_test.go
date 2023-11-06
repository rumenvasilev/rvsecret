package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestPluralize(t *testing.T) {

	Convey("Given a string and a count", t, func() {

		Convey("When the word is 'test'", func() {
			w1 := "test"
			w2 := "tests"

			Convey("If the count is 0", func() {
				w := Pluralize(0, w1, w2)

				Convey("The word should be 'tests'", func() {
					So(w, ShouldEqual, w2)
				})
				Convey("The word should not be 'test'", func() {
					So(w, ShouldNotEqual, w1)
				})
			})

			Convey("If the count is 1", func() {
				w := Pluralize(1, w1, w2)

				Convey("The word should be 'test'", func() {
					So(w, ShouldEqual, w1)
				})
				Convey("The word should not be 'tests'", func() {
					So(w, ShouldNotEqual, w2)
				})
			})

			Convey("If the count is -1", func() {
				w := Pluralize(-1, w1, w2)

				Convey("The word should be 'tests'", func() {
					So(w, ShouldEqual, w2)
				})
				Convey("The word should not be 'test'", func() {
					So(w, ShouldNotEqual, w1)
				})
			})

			Convey("If the count is -0", func() {
				w := Pluralize(-0, w1, w2)

				Convey("The word should be 'tests'", func() {
					So(w, ShouldEqual, w2)
				})
				Convey("The word should not be 'test'", func() {
					So(w, ShouldNotEqual, w1)
				})
			})
		})
	})
}

func TestTruncateString(t *testing.T) {

	Convey("Given a string and a length ", t, func() {
		s := "This is too long"

		Convey("When the string is 'This is too long' and the max length is 10", func() {
			l := 10
			str := TruncateString(s, l)

			Convey("The new string should be 'This is to'", func() {
				So(str, ShouldEqual, "This is to...")
			})

			Convey("The new string should not be 'This is too long' ", func() {
				So(s, ShouldEqual, "This is too long")
			})
		})
	})
}

func TestCleanURLSpaces(t *testing.T) {

	Convey("Given a string", t, func() {

		Convey("When the strings have spaces", func() {
			str := "This "
			result := CleanURLSpaces(str)

			Convey("The spaces should be replaced with dashes", func() {
				for i := range result {
					So(result[i], ShouldEqual, "This-")
				}
			})
			Convey("The spaces should not be replaced with underscores", func() {
				for i := range result {
					So(result[i], ShouldNotEqual, "This_")
				}
			})
			Convey("The spaces should not be left alone", func() {
				for i := range result {
					So(result[i], ShouldNotEqual, "This ")
				}
			})
			Convey("The spaces should not be replaced with \"&#160;\"", func() {
				for i := range result {
					So(result[i], ShouldNotEqual, "This&#160;")
				}
			})
		})

		Convey("When the strings do not have spaces", func() {
			str := "This"
			result := CleanURLSpaces(str)

			Convey("The spaces should be left alone", func() {
				for i := range result {
					So(result[i], ShouldEqual, "This")
				}
			})

			Convey("The string not should contain an extra dash", func() {
				for i := range result {
					So(result[i], ShouldNotEqual, "This-")
				}
			})
			Convey("The string should not contain an extra underscores", func() {
				for i := range result {
					So(result[i], ShouldNotEqual, "This_")
				}
			})
			Convey("The string should not contain an extra \"&#160;\"", func() {
				for i := range result {
					So(result[i], ShouldNotEqual, "This&#160;")
				}
			})
		})
	})
}

func Test_GenerateIDWithLen(t *testing.T) {
	// 1 always generates the same hash
	got := GenerateIDWithLen(1)
	assert.Equal(t, "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", got)

	// 1000 increases randomness so we cannot guarantee we'll have the same hash string output anymore
	got = GenerateIDWithLen(1000)
	assert.NotEmpty(t, got)
}

func Test_GetEntropyInt(t *testing.T) {
	got := GetEntropyInt("a")
	assert.Equal(t, float64(0), got)

	got = GetEntropyInt("something else")
	assert.Equal(t, float64(3.3248629576173565), got)
}

func Test_AppendToSlice(t *testing.T) {
	t.Skip()
}

func Test_AppendIfMissing(t *testing.T) {
	t.Run("append to empty", func(t *testing.T) {
		got := AppendIfMissing([]string{}, "something else")
		assert.Equal(t, []string{"something else"}, got)
		assert.NotEqualValues(t, []string{}, got)
	})

	t.Run("value exist, no change", func(t *testing.T) {
		got := AppendIfMissing([]string{"exist"}, "exist")
		assert.Equal(t, []string{"exist"}, got)
		assert.NotEqualValues(t, []string{"exist", "exist"}, got)
	})

	t.Run("value exist in the middle, no change", func(t *testing.T) {
		got := AppendIfMissing([]string{"exist", "value1", "another"}, "value1")
		assert.Equal(t, []string{"exist", "value1", "another"}, got)
		assert.NotEqualValues(t, []string{"exist", "value1", "another", "value1"}, got)
	})
}

func Test_MergeMaps(t *testing.T) {
	t.Run("nil source", func(t *testing.T) {
		dest := map[string]int{"something": 1}
		MergeMaps(nil, dest)
		assert.Exactly(t, map[string]int{"something": 1}, dest)
	})

	t.Run("nil dest", func(t *testing.T) {
		source := map[string]int{"value": 7}
		var dest map[string]int = nil
		MergeMaps(source, dest)
		assert.Nil(t, dest)
	})

	t.Run("empty, ok", func(t *testing.T) {
		input := make(map[string]int)
		want := make(map[string]int)
		MergeMaps(input, want)
		assert.Exactly(t, make(map[string]int), want)
	})

	t.Run("append only", func(t *testing.T) {
		source := map[string]int{"bla": 7, "test": 1}
		dest := map[string]int{"ok": 10}
		want := map[string]int{"bla": 7, "test": 1, "ok": 10}
		MergeMaps(source, dest)
		assert.Exactly(t, want, dest)
	})

	t.Run("sum values for key", func(t *testing.T) {
		source := map[string]int{"bla": 7, "test": 1}
		dest := map[string]int{"bla": 10}
		want := map[string]int{"bla": 17, "test": 1}
		MergeMaps(source, dest)
		assert.Exactly(t, want, dest)
	})
}
