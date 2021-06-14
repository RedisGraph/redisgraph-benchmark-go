package main

import (
	"math/rand"
	"testing"
)

func Test_processQuery(t *testing.T) {
	type args struct {
		query            string
		randomIntPadding int64
		randomIntMax     int64
	}
	rand.Seed(12345)
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no-replacing", args{"CREATE(n)", 0, 0}, "CREATE(n)"},
		{"no-replacing", args{"ProblemList=[29849199,27107682]", 0, 0}, "ProblemList=[29849199,27107682]"},
		{"no-replacing", args{"ProblemList=[29849199,__rand_int__]", 0, 1}, "ProblemList=[29849199,0]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,__rand_int__]", 0, 1}, "ProblemList=[0,0]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,11]", 0, 1}, "ProblemList=[0,11]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,__rand_int__]", 0, 10}, "ProblemList=[3,4]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,__rand_int__]", -1, 10}, "ProblemList=[7,-1]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processQuery(tt.args.query, tt.args.randomIntPadding, tt.args.randomIntMax); got != tt.want {
				t.Errorf("processQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
