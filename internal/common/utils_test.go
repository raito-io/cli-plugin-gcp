package common

import (
	"fmt"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestIsGoogle400Error(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no error",
			args: args{
				err: nil,
			},
			want: false,
		},
		{
			name: "4xx error",
			args: args{
				err: &googleapi.Error{
					Code: 402,
				},
			},
			want: true,
		},
		{
			name: "403 error (forbidden) - not match",
			args: args{
				err: &googleapi.Error{
					Code: 403,
				},
			},
			want: false,
		},
		{
			name: "wrapped error 4xx - match",
			args: args{
				err: fmt.Errorf("wrapped error: %w", &googleapi.Error{
					Code: 408,
				}),
			},
			want: true,
		}, {
			name: "wrapped error 5xx - not match",
			args: args{
				err: fmt.Errorf("wrapped error: %w", &googleapi.Error{
					Code: 512,
				}),
			},
			want: false,
		}, {
			name: "other error",
			args: args{
				err: fmt.Errorf("wrapped error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGoogle400Error(tt.args.err); got != tt.want {
				t.Errorf("IsGoogle400Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
