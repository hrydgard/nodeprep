// Go implementation of Nodeprep, a special case of Stringprep
// Nodeprep RFC: http://www.ietf.org/rfc/rfc3920.txt
//
// Could easily be generalized to handle any Stringprep
// The public interface is a single function:
//
// func Nodeprep(input string) string
//
// By Henrik Rydgård 2014
//
// BSD Licensed

package nodeprep

import (
  "fmt"
  "strings"
  "strconv"
  "code.google.com/p/go.text/unicode/norm"
)

// Cached lookup tables. Initialized from the strings in nodeprep_tables.go during the first
// time Nodeprep is called. We may want to look into more compact storage...

var mapB1 map[rune]string
var mapB2 map[rune]string
var mapB3 map[rune]string

// If these are in the OUTPUT, the input string is invalid!
var unassigned map[rune]bool

// TODO: Should probably turn into a bitmap or array.
var prohibitMapC1 map[rune]bool

// BIDI stuff. Not used yet.
var charsRandAL map[rune]bool
var charsL map[rune]bool

// Converts a hex number string like 0101 0303 to the corresponding unicode characters and
// concatenates them.
func hexToString(input string) string {
  parts := strings.Split(input, " ")
  str := ""
  for _, p := range(parts) {
    trimmed := strings.TrimSpace(p)
    if len(trimmed) == 0 {
      continue
    }
    val, err := strconv.ParseInt(trimmed, 16, 32)
    if err == nil {
      str += string(rune(val))
    } else {
      fmt.Println(input, "Syntax error in '", trimmed, "':", err)
    }
  }
  return str
}

func parseHexRune(input string) rune {
  val, _ := strconv.ParseInt(input, 16, 32)
  return rune(val)
}

func loadMapTable(str string) map[rune]string {
  table := make(map[rune]string)
  lines := strings.Split(str, "\n")
  for _, line := range(lines) {
    trimmed := strings.TrimSpace(line)
    if len(trimmed) == 0 {
      continue
    }
    parts := strings.Split(trimmed, ";")
    if len(parts) >= 3 {
      from := parseHexRune(strings.TrimSpace(parts[0]))
      to := hexToString(strings.TrimSpace(parts[1]))
      table[from] = to
    }
  }

  return table
}

func loadCharacterSet(str string) map[rune]bool {
  table := make(map[rune]bool)
  lines := strings.Split(str, "\n")
  for _, line := range(lines) {
    trimmed := strings.TrimSpace(line)
    if len(trimmed) == 0 {
      continue
    }
    if trimmed[0] == '-' {
      continue
    }
    a := strings.Split(trimmed, ";")
    parts := strings.Split(a[0], "-")
    if len(parts) == 1 {
      r := parseHexRune(strings.TrimSpace(parts[0]))
      table[r] = true
    } else if len(parts) == 2 {
      from := parseHexRune(strings.TrimSpace(parts[0]))
      to := parseHexRune(strings.TrimSpace(parts[1]))
      for x := from; x <= to; x++ {
        table[x] = true
      }
    }
  }
  return table
}

func remapRunes(input string, mapping map[rune]string) string {
  output := ""
  for _, runeValue := range(input) {
    mapped, present := mapping[runeValue]
    if present {
      output += mapped
    } else {
      output += string(runeValue)
    }
  }
  return output
}

func prohibitRunes(input string, killMap map[rune]bool) string {
  output := ""
  for _, runeValue := range(input) {
    _, present := killMap[runeValue]
    if (!present) {
      output += string(runeValue)
    }
  }
  return output
}

func checkBidiRules(input string) bool {
  if len(input) == 0 {
    return true
  }

  // Section 6 of RFC 3454
  // 1) prohibitions. Already implemented by the prohibit checks
  // 2) "If a string contains any RandALCat character, the string MUST NOT
  //    contain any LCat character"
  // 3) "If a string contains any RandALCat character, a RandALCat
  //    character MUST be the first character of the string, and a
  //    RandALCat character MUST be the last character of the string."

  // Okay, so let's begin with rule 2.
  containsRandAL := false
  containsL := false
  var firstRune, lastRune rune
  firstSet := false
  for _, runeValue := range(input) {
    if !firstSet {
      firstRune = runeValue
      firstSet = true
    }
    _, present := charsRandAL[runeValue] 
    if present {
      containsRandAL = true
    }
    _, present = charsL[runeValue]
    if present {
      containsL = true
    }
    lastRune = runeValue
  }

  if containsRandAL && containsL {
    return false
  }

  // And Rule 3.
  if containsRandAL {
    _, first := charsRandAL[firstRune]
    _, last := charsRandAL[lastRune]
    if first && last {
      return true
    } else {
      return false
    }
  }

  return true
}

func existAnyRunesInSet(input string, bannedMap map[rune]bool) bool {
  for _, runeValue := range(input) {
    _, present := bannedMap[runeValue]
    if present {
      return true
    }
  }
  return false
}

func Nodeprep(input string) (string, error) {
  if mapB1 == nil {
    mapB1 = loadMapTable(mapTableB1str)
    mapB2 = loadMapTable(mapTableB2str)
    loadMapTable(mapTableB3str)  // mapB3 seems unused in the standard
    unassigned = loadCharacterSet(unassignedCodepointsStr)
    prohibitMapC1 = loadCharacterSet(prohibitTableC1str)
    charsRandAL = loadCharacterSet(bidiTableD1str)
    charsL = loadCharacterSet(bidiTableD2str)
  }

  stepDebug := false
  if (stepDebug) {
    fmt.Println("input", input)
  }

  // Step 1: Map
  // According to the RFC we should apply B1 and B2 but not B3. Not sure what's
  // up with that.
  mapped1 := remapRunes(input, mapB1)
  mapped2 := remapRunes(mapped1, mapB2)

  if (stepDebug) {
    fmt.Println("mapped", mapped2)
  }

  // Step 2: Normalize
  // NFKC is specified for the normalization step in RFC 3920
  normalized := string(norm.NFKC.Bytes([]byte(mapped2)))
  if (stepDebug) {
    fmt.Println("normalized", normalized)
  }

  // Step 3: Prohibit runes
  prohibited := prohibitRunes(normalized, prohibitMapC1)
  if (stepDebug) {
    fmt.Println("prohibited", prohibited)
  }

  // Step 4: BIDI checks
  if (!checkBidiRules(prohibited)) {
    return "", fmt.Errorf("String does not comply with BIDI rules")
  }

  // Step 5: Check for unassigned characters
  if existAnyRunesInSet(input, unassigned) {
    return "", fmt.Errorf("String contains characters not in Unicode 3.2")
  }

  return prohibited, nil
}

