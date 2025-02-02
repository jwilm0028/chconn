package cityhash102

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

type TestCase struct {
	key   string
	lower uint64
	upper uint64
}

var testdata = []TestCase{}

func buildData(t *testing.T) {
	f, err := os.Open("testdata/hashs.txt")
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(f)

	var lower uint64
	var upper uint64
	for scanner.Scan() {
		strs := strings.Split(scanner.Text(), ",")

		lower, _ = strconv.ParseUint(strs[1], 16, 64)
		upper, _ = strconv.ParseUint(strs[2], 16, 64)

		testdata = append(testdata, TestCase{strs[0], lower, upper})
	}
}

func check(str string, expected, actual uint64, t *testing.T) {
	if expected != actual {
		t.Errorf("ERROR: %s expected 0x%x but got 0x%x\n", str, expected, actual)
	}
}

func test(str string, lower, upper uint64, t *testing.T) {
	u := CityHash128([]byte(str), uint32(len(str)))

	check(str, lower, u.Lower64(), t)
	check(str, upper, u.Higher64(), t)
}

func Test_Hash(t *testing.T) {
	buildData(t)

	var i int
	for i = 0; i < len(testdata); i++ {
		test(testdata[i].key, testdata[i].lower, testdata[i].upper, t)
	}
}
