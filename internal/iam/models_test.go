package iam

import "testing"

func TestIamBinding_Equals(t *testing.T) {
	type fields struct {
		Member       string
		Role         string
		Resource     string
		ResourceType string
	}
	type args struct {
		b IamBinding
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Equal",
			fields: fields{
				Member:       "member1",
				Role:         "role1",
				Resource:     "resource1",
				ResourceType: "resourceType1",
			},
			args: args{
				b: IamBinding{
					Member:       "member1",
					Role:         "role1",
					Resource:     "resource1",
					ResourceType: "resourceType1",
				},
			},
			want: true,
		},
		{
			name: "Not equal member",
			fields: fields{
				Member:       "member1",
				Role:         "role1",
				Resource:     "resource1",
				ResourceType: "resourceType1",
			},
			args: args{
				b: IamBinding{
					Member:       "member2",
					Role:         "role1",
					Resource:     "resource1",
					ResourceType: "resourceType1",
				},
			},
			want: false,
		},
		{
			name: "Not equal role",
			fields: fields{
				Member:       "member1",
				Role:         "role1",
				Resource:     "resource1",
				ResourceType: "resourceType1",
			},
			args: args{
				b: IamBinding{
					Member:       "member1",
					Role:         "role2",
					Resource:     "resource1",
					ResourceType: "resourceType1",
				},
			},
			want: false,
		},
		{
			name: "Not equal resource",
			fields: fields{
				Member:       "member1",
				Role:         "role1",
				Resource:     "resource1",
				ResourceType: "resourceType1",
			},
			args: args{
				b: IamBinding{
					Member:       "member1",
					Role:         "role1",
					Resource:     "resource2",
					ResourceType: "resourceType1",
				},
			},
			want: false,
		},
		{
			name: "Not equal resource type",
			fields: fields{
				Member:       "member1",
				Role:         "role1",
				Resource:     "resource1",
				ResourceType: "resourceType1",
			},
			args: args{
				b: IamBinding{
					Member:       "member1",
					Role:         "role1",
					Resource:     "resource1",
					ResourceType: "resourceType2",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := IamBinding{
				Member:       tt.fields.Member,
				Role:         tt.fields.Role,
				Resource:     tt.fields.Resource,
				ResourceType: tt.fields.ResourceType,
			}
			if got := a.Equals(tt.args.b); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}
