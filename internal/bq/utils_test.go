package bigquery

import (
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/assert"
)

func Test_getRoleForBQEntity(t *testing.T) {
	type args struct {
		t bigquery.AccessRole
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "owner",
			args: args{
				t: bigquery.OwnerRole,
			},
			want: "roles/bigquery.dataOwner",
		},
		{
			name: "reader",
			args: args{
				t: bigquery.ReaderRole,
			},
			want: "roles/bigquery.dataViewer",
		},
		{
			name: "writer",
			args: args{
				t: bigquery.WriterRole,
			},
			want: "roles/bigquery.dataEditor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getRoleForBQEntity(tt.args.t), "getRoleForBQEntity(%v)", tt.args.t)
		})
	}
}
