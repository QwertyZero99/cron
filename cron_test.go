package cron

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name    string
		args    args
		wantJob Job
		wantErr bool
	}{
		{
			name: "all every",
			args: args{"* * * * *"},
			wantJob: Job{
				Minute:    Field{Every, []int{}},
				Hour:      Field{Every, []int{}},
				Day:       Field{Every, []int{}},
				Month:     Field{Every, []int{}},
				DayOfWeek: Field{Every, []int{}},
				Command:   "",
			},
			wantErr: false,
		},
		{
			name: "all every with command",
			args: args{`* * * * * echo "Hello, world!"`},
			wantJob: Job{
				Minute:    Field{Every, []int{}},
				Hour:      Field{Every, []int{}},
				Day:       Field{Every, []int{}},
				Month:     Field{Every, []int{}},
				DayOfWeek: Field{Every, []int{}},
				Command:   `echo "Hello, world!"`,
			},
			wantErr: false,
		},
		{
			name: "exact",
			args: args{`* 5 * 4 *`},
			wantJob: Job{
				Minute:    Field{Every, []int{}},
				Hour:      Field{Exact, []int{5}},
				Day:       Field{Every, []int{}},
				Month:     Field{Exact, []int{4}},
				DayOfWeek: Field{Every, []int{}},
				Command:   ``,
			},
			wantErr: false,
		},

		{
			name: "step",
			args: args{`* */5 * */4 *`},
			wantJob: Job{
				Minute:    Field{Every, []int{}},
				Hour:      Field{Step, []int{5}},
				Day:       Field{Every, []int{}},
				Month:     Field{Step, []int{4}},
				DayOfWeek: Field{Every, []int{}},
				Command:   ``,
			},
			wantErr: false,
		},
		{
			name: "multiple",
			args: args{`* 5,4 * 3,7,8 *`},
			wantJob: Job{
				Minute:    Field{Every, []int{}},
				Hour:      Field{Multiple, []int{5, 4}},
				Day:       Field{Every, []int{}},
				Month:     Field{Multiple, []int{3, 7, 8}},
				DayOfWeek: Field{Every, []int{}},
				Command:   ``,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJob, err := Parse(tt.args.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotJob, tt.wantJob) {
				t.Errorf("Parse() gotJob = %v, want %v", gotJob, tt.wantJob)
			}
		})
	}
}
