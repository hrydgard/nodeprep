package nodeprep

import (
  _ "fmt"
  "testing"
)

func Test_HexToString(t *testing.T) {
  src := "0041 0042"
  expected := "AB"
  if (hexToString(src) != expected) {
    t.Error(hexToString(src) + " != " + expected)
  }
}

func Test_Nodeprep(t *testing.T) {
  testString := func(t *testing.T, src string, expected string) {
    prepped, err := Nodeprep(src)
    if err != nil {
      t.Error(err)
    }
    if (prepped != expected) {
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
  testFail(t, "ᴮᴵᴳᴮᴵᴿᴰ");
  testFail(t, "BIGᴮIRD");
  testString(t, "aSdFf", "asdff")

  // Test some prohibitions
  testString(t, "S P A C E", "space")
  testString(t, "1&2'3/4:5<6>7@8", "12345678")
}
