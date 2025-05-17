// Package cron helps parse cron job expressions, and perform common operations on them.
// It is not official and does not necessarily apply to any standards (if they exist IDK).
package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// fieldType represents a type of Field e.g. Every for *, and Multiple for 1,3.
// Includes Exact, Every, Multiple, Range, and Step (Any is available but not supported yet).
type fieldType int

const (
	Exact fieldType = iota
	Every
	Multiple
	Range
	Step
	Any // TODO: Support for quartz-style "?" maybe
)

// Field represents a time field in a cron-job. Not meant for common use.
type Field struct {
	Type   fieldType
	Values []int
}

// Field.String parses back into a cron-expression field.
func (f Field) String() string {
	switch f.Type {
	case Exact:
		if len(f.Values) != 1 {
			panic(fmt.Sprintf("Exact type requires 1 value; got %d", len(f.Values)))
		}
		return strconv.Itoa(f.Values[0])
	case Every:
		return "*"
	case Multiple:
		if len(f.Values) == 0 {
			panic("Multiple type requires at least one value")
		}
		multipleStrings := make([]string, len(f.Values))
		for i, val := range f.Values {
			multipleStrings[i] = strconv.Itoa(val)
		}
		return strings.Join(multipleStrings, ",")
	case Range:
		if len(f.Values) != 2 {
			panic("Range type requires exactly 2 values")
		}
		return fmt.Sprintf("%d-%d", f.Values[0], f.Values[1])
	case Step:
		if len(f.Values) != 1 {
			panic("Step type requires 1 value")
		}
		return fmt.Sprintf("*/%d", f.Values[0])
	case Any:
		return "?"
	default:
		return "???"
	}
}

// ToEnglish turns a job into an english, human-readable string. TODO improve this function.
func (job Job) ToEnglish() string {
	return ""
}

func (f Field) check(val int) bool {
	switch f.Type {
	case Exact:
		return len(f.Values) == 1 && f.Values[0] == val

	case Every:
		return true

	case Multiple:
		for _, v := range f.Values {
			if v == val {
				return true
			}
		}
		return false

	case Range:
		if len(f.Values) != 2 {
			return false
		}
		return val >= f.Values[0] && val <= f.Values[1]

	case Step:
		if len(f.Values) != 1 || f.Values[0] <= 0 {
			return false
		}
		return val%f.Values[0] == 0

	case Any:
		// TODO: Not implemented yet
		return false

	default:
		return false
	}
}

// Job represents a cron-job and contains each time field, and a task as a string to complete.
type Job struct {
	Minute    Field
	Hour      Field
	Day       Field
	Month     Field
	DayOfWeek Field
	Task      string
}

// Job.String parses back into a cron-expression.
func (job Job) String() string {
	return fmt.Sprintf(
		"%s %s %s %s %s %s",
		job.Minute, job.Hour, job.Day, job.Month, job.DayOfWeek, job.Task,
	)
}

// Check should tell you if the cron job applies to a time
func (job Job) Check(t time.Time) bool {
	return job.Minute.check(t.Minute()) &&
		job.Hour.check(t.Hour()) &&
		job.Day.check(t.Day()) &&
		job.Month.check(int(t.Month())) && // time.Month -> int
		job.DayOfWeek.check(int(t.Weekday())) // time.Weekday -> int
}

// IsNow is a wrapper for Job.Check(time.Now())
func (job Job) IsNow() bool {
	return job.Check(time.Now())
}

// ==Parsing==

// parseField parses a field into a Field struct
func parseField(fieldString string) (Field, error) {
	s := strings.TrimSpace(fieldString)

	switch {
	case s == "*":
		return Field{Type: Every, Values: []int{}}, nil

	case strings.HasPrefix(s, "*/"):
		val, err := strconv.Atoi(s[2:])
		if err != nil {
			return Field{}, err
		}
		return Field{Type: Step, Values: []int{val}}, nil

	case strings.Contains(s, ","):
		parts := strings.Split(s, ",")
		values := make([]int, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.Atoi(part)
			if err != nil {
				return Field{}, err
			}
			values = append(values, val)
		}
		return Field{Type: Multiple, Values: values}, nil

	case strings.Contains(s, "-"):
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return Field{}, fmt.Errorf("invalid range format: %s", s)
		}
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return Field{}, err
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return Field{}, err
		}
		return Field{Type: Range, Values: []int{start, end}}, nil

	case s == "?":
		return Field{Type: Any}, nil

	default:
		val, err := strconv.Atoi(s)
		if err != nil {
			return Field{}, err
		}
		return Field{Type: Exact, Values: []int{val}}, nil
	}
}

// Parse parses a string in cron expression format (e.g. '* */5 5 * * echo "Hello, world"') into a `Job` struct.
func Parse(expression string) (Job, error) {
	parts := strings.Fields(expression)
	if len(parts) < 5 {
		return Job{}, fmt.Errorf("expected at least 6 fields (5 time + task), got %d", len(parts))
	}

	timeFields := parts[:5]
	task := strings.Join(parts[5:], " ")

	job := Job{}
	var err error

	if job.Minute, err = parseField(timeFields[0]); err != nil {
		return Job{}, fmt.Errorf("minute field: %w", err)
	}
	if job.Hour, err = parseField(timeFields[1]); err != nil {
		return Job{}, fmt.Errorf("hour field: %w", err)
	}
	if job.Day, err = parseField(timeFields[2]); err != nil {
		return Job{}, fmt.Errorf("day field: %w", err)
	}
	if job.Month, err = parseField(timeFields[3]); err != nil {
		return Job{}, fmt.Errorf("month field: %w", err)
	}
	if job.DayOfWeek, err = parseField(timeFields[4]); err != nil {
		return Job{}, fmt.Errorf("dayOfWeek field: %w", err)
	}

	job.Task = task
	return job, nil
}
