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

func pythonNodeprep(s []string) (result []string, legal []bool, err error) {
	cmd := exec.Command("python", filepath.Join("python", "convert.py"))
	cmd.Stdin = bytes.NewBufferString(strings.Join(s, "\n"))
	combined, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result = make([]string, len(s))
	legal = make([]bool, len(s))
	for i, prepped := range strings.Split(string(combined), "\n") {
		if prepped != "ILLEGAL" {
			legal[i] = true
			result[i] = prepped
		}
	}
	return
}

const (
	maxRune      = 0x0010FFFF
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
	practicalMax = maxRune - (surrogateMax - surrogateMin)
)

func randRune() (result rune) {
	result = rune(rand.Intn(practicalMax))
	if surrogateMin <= result && result <= surrogateMax {
		result += (surrogateMax - surrogateMin)
	}
	return
}

func testStrings(t *testing.T, strings []string) {
	prep1, legal1, err := pythonNodeprep(strings)
	if err != nil {
		t.Fatalf("%v", err)
	}
	prep2, legal2, err := pythonNodeprep(prep1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	errs := 0
	correct := 0
	bad, err := os.Create("badpreps")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer bad.Close()
	for i, s := range strings {
		goPrep, err := Nodeprep(s)
		if err == nil {
			if !legal1[i] || prep1[i] != prep2[i] {
				t.Errorf("Python said %#v was illegal to prep (it produced %#v, %v and %#v, %v), but Go prepping it produced %#v", strings[i], prep1[i], legal1[i], prep2[i], legal2[i], goPrep)
				fmt.Fprintln(bad, strings[i])
				errs++
			} else if prep1[i] != goPrep {
				t.Errorf("Python prepping %#v produced %#v, but Go prepping it produced %#v", strings[i], prep1[i], goPrep)
				fmt.Fprintln(bad, strings[i])
				errs++
			} else {
				correct++
			}
		} else {
			if legal1[i] && prep1[i] == prep2[i] {
				t.Errorf("Python prepping %#v produced %#v, but was unable to Go prep it: %v", strings[i], prep1[i], err)
				fmt.Fprintln(bad, strings[i])
				errs++
			}
		}
	}
	if errs > 0 {
		t.Errorf("%v/%v strings badly encoded (but %v non zero strings correctly encoded :)", errs, len(strings), correct)
	}
}

func Test_RandomCodePoints(t *testing.T) {
	n := 100000
	randomStrings := make([]string, n)
	for i := 0; i < n; i++ {
		buf := make([]rune, 3)
		buf[0] = randRune()
		buf[1] = randRune()
		buf[2] = randRune()
		if !utf8.ValidString(string(buf)) {
			t.Fatalf("Unable to create valid UTF8 string: %#v", string(buf))
		}
		randomStrings[i] = string(buf)
	}
	fmt.Printf("Generated %v random and valid 3 char UTF8 strings\n", n)
	testStrings(t, randomStrings)
}
