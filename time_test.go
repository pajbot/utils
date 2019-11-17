package utils

import (
	"testing"
	"time"
)

const timeFormat = "Mon Jan 2 15:04:05.000"

func testRelTime(t *testing.T, t1, t2 time.Time, expectedResult string) {
	result := RelTime(t1, t2)
	if result != expectedResult {
		t.Error("Got '" + result + "', expected '" + expectedResult + "'")
	}

	var diff time.Duration
	if t1.After(t2) {
		diff = t1.Sub(t2)
	} else {
		diff = t2.Sub(t1)
	}

	result = DurationString(diff)
	if result != expectedResult {
		t.Error("Got '" + result + "', expected '" + expectedResult + "'")
	}
}

func testCustomRelTime(t *testing.T, t1, t2 time.Time, numParts int, glue, expectedResult string) {
	result := CustomRelTime(t1, t2, numParts, glue)
	if result != expectedResult {
		t.Error("Got '" + result + "', expected '" + expectedResult + "'")
	}

	var diff time.Duration
	if t1.After(t2) {
		diff = t1.Sub(t2)
	} else {
		diff = t2.Sub(t1)
	}
	result = CustomDurationString(diff, numParts, glue)
	if result != expectedResult {
		t.Error("Got '" + result + "', expected '" + expectedResult + "'")
	}
}

type timeTest1 struct {
	TimeStringA string
	TimeStringB string

	ExpectedDiffString string
}

func TestRelTime(t *testing.T) {
	var t1 time.Time
	var t2 time.Time
	var err error

	var tests = []timeTest1{
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:00:05.000", "4 minutes"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:00:04.000", "4 minutes 1 second"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:00:03.000", "4 minutes 2 seconds"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:04:03.000", "2 seconds"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:04:04.000", "1 second"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:04:04.001", "now"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 15:04:04.999", "now"},
		{"Mon Jan 3 15:04:04.999", "Mon Jan 3 15:04:05.000", "now"},
		{"Mon Jan 3 15:00:03.000", "Mon Jan 3 15:04:05.000", "4 minutes 2 seconds"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 2 15:04:05.000", "1 day"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 1 15:04:05.000", "2 days"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 14:04:05.000", "1 hour"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 3 13:04:05.000", "2 hours"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 10 15:04:05.000", "1 week"},
		{"Mon Jan 3 15:04:05.000", "Mon Jan 17 15:04:05.000", "2 weeks"},
		{"Mon Sep 1 15:04:05.000", "Mon Oct 1 15:04:05.000", "1 month"},
		{"Mon Jan 3 15:04:05.000", "Mon Mar 3 15:04:05.000", "2 months"},
	}

	for _, test := range tests {
		t1, err = time.Parse(timeFormat, test.TimeStringA)
		if err != nil {
			t.Fatal(err)
		}
		t2, err = time.Parse(timeFormat, test.TimeStringB)
		if err != nil {
			t.Fatal(err)
		}
		testRelTime(t, t1, t2, test.ExpectedDiffString)
	}
}

func TestCustomRelTime(t *testing.T) {
	var t1 time.Time
	var t2 time.Time

	t1, _ = time.Parse(timeFormat, "Mon Jan 3 15:04:05.000")
	t2, _ = time.Parse(timeFormat, "Mon Jan 3 15:00:05.000")
	testCustomRelTime(t, t1, t2, 1, " ", "4 minutes")

	t1, _ = time.Parse(timeFormat, "Mon Jan 3 15:04:05.000")
	t2, _ = time.Parse(timeFormat, "Mon Jan 3 15:00:02.000")
	testCustomRelTime(t, t1, t2, 1, " ", "4 minutes")

	t1, _ = time.Parse(timeFormat, "Mon Jan 3 15:04:05.000")
	t2, _ = time.Parse(timeFormat, "Mon Jan 3 15:00:02.000")
	testCustomRelTime(t, t1, t2, 2, " ", "4 minutes 3 seconds")

	t1, _ = time.Parse(timeFormat, "Mon Jan 3 15:04:05.000")
	t2, _ = time.Parse(timeFormat, "Mon Jan 3 15:00:02.000")
	testCustomRelTime(t, t1, t2, 3, " ", "4 minutes 3 seconds")

	t1, _ = time.Parse(timeFormat, "Mon Jan 3 15:04:05.000")
	t2, _ = time.Parse(timeFormat, "Mon Jan 3 15:00:02.001")
	testCustomRelTime(t, t1, t2, 3, " ", "4 minutes 2 seconds")

	t1, _ = time.Parse(timeFormat, "Mon Jan 3 15:04:05.000")
	t2, _ = time.Parse(timeFormat, "Mon Jan 2 14:03:04.000")
	testCustomRelTime(t, t1, t2, 1, " ", "1 day")
	testCustomRelTime(t, t1, t2, 2, " ", "1 day 1 hour")
	testCustomRelTime(t, t1, t2, 3, " ", "1 day 1 hour 1 minute")
	testCustomRelTime(t, t1, t2, 4, " ", "1 day 1 hour 1 minute 1 second")

	testCustomRelTime(t, t1, t2, 1, ", ", "1 day")
	testCustomRelTime(t, t1, t2, 2, ", ", "1 day, 1 hour")
	testCustomRelTime(t, t1, t2, 3, ", ", "1 day, 1 hour, 1 minute")
	testCustomRelTime(t, t1, t2, 4, ", ", "1 day, 1 hour, 1 minute, 1 second")
}

type parseDurationTest struct {
	durationString string

	expectedDuration time.Duration
	expectedError    error
}

func TestParseDuration(t *testing.T) {
	tests := []parseDurationTest{
		{"1s", 1 * time.Second, nil},
		{"1m", 1 * time.Minute, nil},
		{"1h", 1 * time.Hour, nil},
		{"1d", 24 * time.Hour, nil},
		{"-1s", -1 * time.Second, nil},
		{"-1m", -1 * time.Minute, nil},
		{"-1h", -1 * time.Hour, nil},
		{"-1d", -24 * time.Hour, nil},
		{"1h1m", 1*time.Hour + 1*time.Minute, nil},
		{"-1h", -1 * time.Hour, nil},
		{"0.5h", 30 * time.Minute, nil},
		{"0.10h", 6 * time.Minute, nil},
		{"0", 0, nil},
		{"1x", 0, errUnknownUnit},
		{".s", 0, errA},
		{"-.s", 0, errA},
	}

	for _, test := range tests {
		dur, err := ParseDuration(test.durationString)
		if test.expectedError != nil {
			if test.expectedError != err {
				t.Fatalf("wrong error, got %s but expected %s", err, test.expectedError)
			}

			if dur != 0 {
				t.Fatalf("expected 0 duration, got %d", dur)
			}
		} else if test.expectedDuration != dur {
			t.Fatalf("wrong duration: got %s but expected %s", dur, test.expectedDuration)
		}
	}
}
