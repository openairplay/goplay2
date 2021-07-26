//+build linux

package codec

import "testing"

func Test_dbToLinearVolume(t *testing.T) {
	type args struct {
		volume float64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "-144db",
			args: args{
				volume: -144.0,
			},
			want: 261,
		},
		{
			name: "-30db",
			args: args{
				volume: -30.0,
			},
			want: 20724,
		},
		{
			name: "0db",
			args: args{
				volume: .0,
			},
			want: 65536,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dbToLinearVolume(tt.args.volume); got != tt.want {
				t.Errorf("dbToLinearVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
