package cron

import (
	"errors"
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
	Any // TODO from quartz
)

type Field struct {
	Type   fieldType
	Values []int
}

func (f Field) String() string {
	switch f.Type {
	case Exact:
		if len(f.Values) != 1 {
			panic(fmt.Sprintf("incorrect amount of values in exact Field; want (1); have(%d)", len(f.Values)))
		}
		return strconv.Itoa(f.Values[0])
	case Every:
		return "*"
	case Multiple:
		if len(f.Values) <= 0 {
			panic(fmt.Sprintf("incorrect amount of values in multiple Field; want (len>0); have(%d)", len(f.Values)))
		}
		var sb strings.Builder
		for _, value := range f.Values {
			sb.WriteString(strconv.Itoa(value) + ",")
		}
		str := sb.String()
		return str[:len(str)-2] // Return without the final comma
	case Range:
		return strconv.Itoa(f.Values[0]) + "-" + strconv.Itoa(f.Values[1])
	case Step:
		if len(f.Values) != 1 {
			panic(fmt.Sprintf("incorrect amount of values in step Field; want (1); have(%d)", len(f.Values)))
		}
		return "*/" + strconv.Itoa(f.Values[0])
	case Any:
		return "?"
	}
	return "???"
}

type Job struct {
	Minute    Field
	Hour      Field
	Day       Field
	Month     Field
	DayOfWeek Field
	Command   string
}

func (j Job) String() (s string) {
	var sb strings.Builder
	sb.WriteString(j.Minute.String() + " ")
	sb.WriteString(j.Hour.String() + " ")
	sb.WriteString(j.Day.String() + " ")
	sb.WriteString(j.Month.String() + " ")
	sb.WriteString(j.DayOfWeek.String() + " ")
	sb.WriteString(j.Command)
	return strings.TrimSpace(sb.String())
}

func parseField(fieldString string) (field Field, err error) {
	fieldStringTrimmed := strings.TrimSpace(fieldString)
	if fieldStringTrimmed == "*" {
		return Field{Type: Every, Values: []int{}}, err
	}
	if cut, found := strings.CutPrefix(fieldStringTrimmed, "*/"); found {
		val, convErr := strconv.Atoi(cut)
		if convErr != nil {
			return Field{}, convErr
		}
		return Field{Type: Step, Values: []int{val}}, err
	}

	if split := strings.Split(fieldStringTrimmed, ","); len(split) > 1 {
		for _, s := range split {
			val, convErr := strconv.Atoi(s)
			if convErr != nil {
				return Field{}, convErr
			}
			field.Values = append(field.Values, val)
		}
		field.Type = Multiple
		return field, err
	}

	if split := strings.Split(fieldStringTrimmed, "-"); len(split) == 2 {
		field.Values[0], err = strconv.Atoi(split[0])
		if err != nil {
			return Field{}, err
		}
		field.Values[1], err = strconv.Atoi(split[1])
		if err != nil {
			return Field{}, err
		}

		field.Type = Multiple
		return field, err
	}

	// Exact
	val, convErr := strconv.Atoi(fieldStringTrimmed)
	if convErr != nil {
		return Field{}, convErr
	}
	return Field{Type: Exact, Values: []int{val}}, err
}

func Parse(expression string) (job Job, err error) {
	stringFields := strings.Fields(expression)
	commandArgs := stringFields[5:]
	command := strings.Join(commandArgs, " ")
	stringFields = stringFields[:5]
	if len(stringFields) != 6 && len(stringFields) != 5 {
		return Job{}, errors.New(fmt.Sprintf("expected 5 or 6 fields, got %d (including command)", len(stringFields)+1)) // +1 is for the command
	}
	job.Minute, err = parseField(stringFields[0])
	if err != nil {
		return Job{}, err
	}
	job.Hour, err = parseField(stringFields[1])
	if err != nil {
		return Job{}, err
	}
	job.Day, err = parseField(stringFields[2])
	if err != nil {
		return Job{}, err
	}
	job.Month, err = parseField(stringFields[3])
	if err != nil {
		return Job{}, err
	}
	job.DayOfWeek, err = parseField(stringFields[4])
	if err != nil {
		return Job{}, err
	}
	job.Command = command
	return job, err
}

func splitLastWord(input string) (string, string) {
	lastSpace := strings.LastIndex(input, " ")
	if lastSpace == -1 {
		// No space found, return empty first part and entire input as second
		return "", input
	}
	first := input[:lastSpace]
	second := input[lastSpace+1:]
	return first, second
}
