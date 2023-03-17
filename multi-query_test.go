package main

import "testing"

func Test_sample(t *testing.T) {
	type args struct {
		cdf []float32
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "last bucket", args: args{[]float32{0.01, 0.99}}, want: 1},
		{name: "last bucket with high likelyhood", args: args{[]float32{0.00001, 0.99}}, want: 1},
		{name: "single bucket", args: args{[]float32{0.99}}, want: 0},
		{name: "after bucket", args: args{[]float32{0.00000001}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sample(tt.args.cdf); got != tt.want {
				t.Errorf("sample() = %v, want %v", got, tt.want)
			}
		})
	}
}
