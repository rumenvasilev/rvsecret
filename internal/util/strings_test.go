package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
