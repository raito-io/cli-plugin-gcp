package roles

import (
	"reflect"
	"testing"

	ds "github.com/raito-io/cli/base/data_source"
)

func Test_RoleToDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Owner",
			input:    "roles/owner",
			expected: "Owner",
		},
		{
			name:     "Bigquery Dataviewer",
			input:    "roles/bigquery.dataViewer",
			expected: "Bigquery Dataviewer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoleToDisplayName(tt.input); got != tt.expected {
				t.Errorf("RoleToDisplayName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGcpRole_ToDataObjectTypePermission(t *testing.T) {
	type fields struct {
		Name                   string
		Description            string
		GlobalPermissions      map[Service][]string
		UsageGlobalPermissions map[Service][]string
	}
	type args struct {
		service Service
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ds.DataObjectTypePermission
	}{
		{
			name: "GcpRole without global permissions",
			fields: fields{
				Name:                   "roles/bigquery.user",
				Description:            "Some description for this role",
				GlobalPermissions:      nil,
				UsageGlobalPermissions: nil,
			},
			args: args{
				service: ServiceBigQuery,
			},
			want: &ds.DataObjectTypePermission{
				Permission:             "roles/bigquery.user",
				Description:            "Some description for this role",
				GlobalPermissions:      nil,
				UsageGlobalPermissions: nil,
			},
		},
		{
			name: "GcpRole with non-applicable global permissions",
			fields: fields{
				Name:                   "roles/bigquery.user",
				Description:            "Some description for this role",
				GlobalPermissions:      map[Service][]string{ServiceBigQuery: {ds.Read}},
				UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Write}},
			},
			args: args{
				service: ServiceGcp,
			},
			want: &ds.DataObjectTypePermission{
				Permission:             "roles/bigquery.user",
				Description:            "Some description for this role",
				GlobalPermissions:      nil,
				UsageGlobalPermissions: nil,
			},
		},
		{
			name: "GcpRole with applicable global permissions",
			fields: fields{
				Name:                   "roles/bigquery.user",
				Description:            "Some description for this role",
				GlobalPermissions:      map[Service][]string{ServiceBigQuery: {ds.Read}},
				UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Write}},
			},
			args: args{
				service: ServiceBigQuery,
			},
			want: &ds.DataObjectTypePermission{
				Permission:             "roles/bigquery.user",
				Description:            "Some description for this role",
				GlobalPermissions:      []string{ds.Read},
				UsageGlobalPermissions: []string{ds.Write},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &GcpRole{
				Name:                   tt.fields.Name,
				Description:            tt.fields.Description,
				GlobalPermissions:      tt.fields.GlobalPermissions,
				UsageGlobalPermissions: tt.fields.UsageGlobalPermissions,
			}
			if got := r.ToDataObjectTypePermission(tt.args.service); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDataObjectTypePermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
