package common

import (
	"testing"

	"github.com/raito-io/cli/base/data_source"
)

func TestShouldHandle(t *testing.T) {
	type args struct {
		fullName string
		config   *data_source.DataSourceSyncConfig
	}
	tests := []struct {
		name    string
		args    args
		wantRet bool
	}{
		{
			name:    "Should handle - no dataObjectParent",
			args:    args{fullName: "projects.raito-io.datasets.test", config: &data_source.DataSourceSyncConfig{DataObjectParent: ""}},
			wantRet: true,
		},
		{
			name:    "Should handle - fullname under dataObjectParent",
			args:    args{fullName: "projects.raito-io.datasets.test", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io"}},
			wantRet: true,
		},
		{
			name:    "Should not handle handle - fullname equal to dataObjectParent",
			args:    args{fullName: "projects.raito-io.datasets.test", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io.datasets.test"}},
			wantRet: false,
		},
		{
			name:    "Should not handle - if fullname not match",
			args:    args{fullName: "projects.raito-io.datasets2.test2", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io.datasets3"}},
			wantRet: false,
		},
		{
			name:    "Should not handle - if exclude match",
			args:    args{fullName: "projects.raito-io.datasets.test.1234", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io.datasets", DataObjectExcludes: []string{"test", "1234"}}},
			wantRet: false,
		},
		{
			name:    "Should handle - no exclude matching",
			args:    args{fullName: "projects.raito-io.datasets.test.1234", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io", DataObjectExcludes: []string{"datasets2.test", "datasets2.1234"}}},
			wantRet: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRet := ShouldHandle(tt.args.fullName, tt.args.config); gotRet != tt.wantRet {
				t.Errorf("ShouldHandle() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestShouldGoInto(t *testing.T) {
	type args struct {
		fullName string
		config   *data_source.DataSourceSyncConfig
	}
	tests := []struct {
		name    string
		args    args
		wantRet bool
	}{
		{
			name:    "Should go into - no dataObjectParent",
			args:    args{fullName: "projects.raito-io.datasets.test", config: &data_source.DataSourceSyncConfig{DataObjectParent: ""}},
			wantRet: true,
		},
		{
			name:    "Should go into - fullname under dataObjectParent",
			args:    args{fullName: "projects.raito-io.datasets.test", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io"}},
			wantRet: true,
		},
		{
			name:    "Should not go into - dataObjectParent under fullname",
			args:    args{fullName: "projects.raito-io.datasets", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io.datasets.test"}},
			wantRet: true,
		},
		{
			name:    "Should not go into - if fullname not match",
			args:    args{fullName: "projects.raito-io.datasets2.test2", config: &data_source.DataSourceSyncConfig{DataObjectParent: "projects.raito-io.datasets3"}},
			wantRet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRet := ShouldGoInto(tt.args.fullName, tt.args.config); gotRet != tt.wantRet {
				t.Errorf("ShouldGoInto() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
