package nodeprep

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Test_HexToString(t *testing.T) {
	src := "0041 0042"
	expected := "AB"
	if hexToString(src) != expected {
		t.Error(hexToString(src) + " != " + expected)
	}
}

func Test_Nodeprep(t *testing.T) {
	testString := func(t *testing.T, src string, expected string) {
		prepped, err := Nodeprep(src)
		if err != nil {
			t.Error(err)
		}
		if prepped != expected {
			t.Error("Nodeprep(" + src + ") (" + prepped + ") != " + expected)
		}
	}

	testFail := func(t *testing.T, src string) {
		_, err := Nodeprep(src)
		if err == nil {
			t.Error("Should fail!")
		}
	}

	// Nodeprep will fail any strings outside Unicode 3.2. Otherwise it is not idempotent,
	// as the following example would demonstrate:
	// testString(t, "ᴮᴵᴳᴮᴵᴿᴰ", "BIGBIRD")
	// testString(t, "BIGBIRD", "bigbird")
	// This was noted by Spotify in http://labs.spotify.com/2013/06/18/creative-usernames/#comments
	// Check that we fail:
	testFail(t, "ᴮᴵᴳᴮᴵᴿᴰ")
	testFail(t, "BIGᴮIRD")
	testString(t, "aSdFf", "asdff")

	// Test some prohibitions
	testString(t, "S P A C E", "space")
	testString(t, "1&2'3/4:5<6>7@8", "12345678")
}

func pythonNodeprep(s []string) (result []string, err error) {
	cmd := exec.Command("python", filepath.Join("python", "convert.py"))
	cmd.Stdin = bytes.NewBufferString(strings.Join(s, "\n"))
	combined, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result = make([]string, len(s))
	for i, prepped := range strings.Split(string(combined), "\n") {
		if prepped != "ILLEGAL" {
			result[i] = prepped
		}
	}
	return
}

var hits = 0
var misses = 0

func randRune() (result rune) {
	result = rune(int64(rand.Int())<<32 + int64(rand.Int()))
	for !utf8.ValidRune(result) {
		misses++
		result = rune(int64(rand.Int())<<32 + int64(rand.Int()))
	}
	hits++
	return
}

func Test_RandomStrings(t *testing.T) {
	n := 10000
	length := 10
	randomStrings := make([]string, n)
	for i := 0; i < n; i++ {
		l := 1 + rand.Int()%length
		buf := make([]rune, l)
		for j := 0; j < l; j++ {
			buf[j] = randRune()
		}
		if !utf8.ValidString(string(buf)) {
			t.Fatalf("Unable to create valid UTF8 string: %#v", string(buf))
		}
		randomStrings[i] = string(buf)
		fmt.Printf("\rGenerated %v random and valid UTF8 strings ", i)
	}
	fmt.Printf("\nFound %v/%v valid code points (on average %v of the possible runes are valid code points)\n", hits, hits+misses, float64(hits)/float64(hits+misses))
	prep1, err := pythonNodeprep(randomStrings)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for i := 0; i < n; i++ {
		if prep1[i] != "" {
			fmt.Printf("%#v => %#v\n", randomStrings[i], prep1[i])
		}
	}
	prep2, err := pythonNodeprep(prep1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for i := 0; i < n; i++ {
		if prep1[i] != "" {
			fmt.Printf("%#v => %#v\n", prep1[i], prep2[i])
		}
	}
	errs := 0
	correct := 0
	bad, err := os.Create("badpreps")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer bad.Close()
	for i, s := range randomStrings {
		goPrep1, err := Nodeprep(s)
		if err == nil {
			if prep1[i] != goPrep1 {
				t.Errorf("Python prepping %#v produced %#v, but Go prepping it produced %#v", randomStrings[i], prep1[i], goPrep1)
				fmt.Fprintln(bad, randomStrings[i])
				errs++
			}
			goPrep2, err := Nodeprep(goPrep1)
			if err == nil {
				if prep2[i] != goPrep2 {
					t.Errorf("Python prepping %#v twice produced %#v, but Go prepping it twice produced %#v", randomStrings[i], prep2[i], goPrep2)
					fmt.Fprintln(bad, randomStrings[i])
					errs++
				} else {
					correct++
				}
			} else {
				if prep2[i] != "" {
					t.Errorf("Python prepping %#v twice produced %#v, but was unable to Go prep it twice: %v", randomStrings[i], prep2[i], err)
					fmt.Fprintln(bad, randomStrings[i])
					errs++
				}
			}
		} else {
			if prep1[i] != "" {
				t.Errorf("Python prepping %#v produced %#v, but was unable to Go prep it: %v", randomStrings[i], prep1[i], err)
				fmt.Fprintln(bad, randomStrings[i])
				errs++
			}
		}
	}
	if errs > 0 {
		t.Errorf("%v/%v strings badly encoded (but %v non zero strings correctly encoded :)", errs, n, correct)
	}
}
