package main

import "testing"

func Test_toTime(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "sec", args: args{str: "2.2378"}, want: "00:00:02.237"},
		{name: "min", args: args{str: "62.2378"}, want: "00:01:02.237"},
		{name: "hour", args: args{str: "3662.2378"}, want: "01:01:02.237"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toTime(tt.args.str); got != tt.want {
				t.Errorf("toTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
