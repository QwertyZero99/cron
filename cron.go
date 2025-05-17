package cron

import (
	"fmt"
	"strconv"
	"strings"
)

type fieldType int

const (
	Exact fieldType = iota
	Every
	Multiple
	Range
	Step
	Any // TODO: Support for quartz-style "?"
)

type Field struct {
	Type   fieldType
	Values []int
}

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
		strs := make([]string, len(f.Values))
		for i, val := range f.Values {
			strs[i] = strconv.Itoa(val)
		}
		return strings.Join(strs, ",")
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

type Job struct {
	Minute    Field
	Hour      Field
	Day       Field
	Month     Field
	DayOfWeek Field
	Command   string
}

func (j Job) String() string {
	return fmt.Sprintf(
		"%s %s %s %s %s %s",
		j.Minute, j.Hour, j.Day, j.Month, j.DayOfWeek, j.Command,
	)
}

func parseField(fieldString string) (Field, error) {
	s := strings.TrimSpace(fieldString)

	switch {
	case s == "*":
		return Field{Type: Every}, nil

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

func Parse(expression string) (Job, error) {
	parts := strings.Fields(expression)
	if len(parts) < 6 {
		return Job{}, fmt.Errorf("expected at least 6 fields (5 time + command), got %d", len(parts))
	}

	timeFields := parts[:5]
	command := strings.Join(parts[5:], " ")

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

	job.Command = command
	return job, nil
}

func splitLastWord(input string) (string, string) {
	if idx := strings.LastIndex(input, " "); idx != -1 {
		return input[:idx], input[idx+1:]
	}
	return "", input
}
