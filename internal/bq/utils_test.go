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

func Test_validSqlName(t *testing.T) {
	type args struct {
		originalName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid name",
			args: args{
				originalName: "valid_name",
			},
			want: "valid_name",
		},
		{
			name: "replace all spaces",
			args: args{
				originalName: "valid name",
			},
			want: "valid_name",
		},
		{
			name: "replace all non alphanumeric",
			args: args{
				originalName: "valid-name🤦‍^$()#@12+3!",
			},
			want: "valid_name12_3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validSqlName(tt.args.originalName), "validSqlName(%v)", tt.args.originalName)
		})
	}
}
