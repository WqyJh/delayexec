package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestRegex(t *testing.T) {
	str := "test space"
	re := regexp.MustCompile(`\s+`)
	if re.Match([]byte(str)) {
		str = fmt.Sprintf(`"%s"`, str)
	}
	if str != `"test space"` {
		t.Fatalf("space regex failed")
	}
	t.Logf("str: %s", str)

	str = strings.Join([]string{str}, " ")
	if str != `"test space"` {
		t.Fatalf("space regex failed")
	}
	t.Logf("str: %s", str)
}

func TestTimeParse(t *testing.T) {
	ts := "2022-04-05 04:45:00"
	timestamp := 1649133900
	result, err := time.Parse(TimeFormat, ts)
	if err != nil {
		t.Fatalf("parse time failed: %s", err.Error())
	}
	if result.Unix() != int64(timestamp) {
		t.Fatalf("time not equal")
	}
	t.Logf("result: %d timestamp: %d", result.Unix(), timestamp)
}

func TestTimeToCron(t *testing.T) {
	ts := "2022-04-05 04:45:00"
	result, err := time.Parse(TimeFormat, ts)
	if err != nil {
		t.Fatalf("parse time failed: %s", err.Error())
	}
	cron := TimeToCron(result)
	t.Logf(`cron: "%s"`, cron)
	if cron != "45 4 5 4 *" {
		t.Fatalf("cron not equal")
	}
}
