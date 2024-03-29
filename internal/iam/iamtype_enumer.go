// Code generated by "enumer -gqlgen -type=IamType"; DO NOT EDIT.

package iam

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const _IamTypeName = "ProjectFolderOrganizationGSuiteService"

var _IamTypeIndex = [...]uint8{0, 7, 13, 25, 31, 38}

const _IamTypeLowerName = "projectfolderorganizationgsuiteservice"

func (i IamType) String() string {
	if i < 0 || i >= IamType(len(_IamTypeIndex)-1) {
		return fmt.Sprintf("IamType(%d)", i)
	}
	return _IamTypeName[_IamTypeIndex[i]:_IamTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _IamTypeNoOp() {
	var x [1]struct{}
	_ = x[Project-(0)]
	_ = x[Folder-(1)]
	_ = x[Organization-(2)]
	_ = x[GSuite-(3)]
	_ = x[Service-(4)]
}

var _IamTypeValues = []IamType{Project, Folder, Organization, GSuite, Service}

var _IamTypeNameToValueMap = map[string]IamType{
	_IamTypeName[0:7]:        Project,
	_IamTypeLowerName[0:7]:   Project,
	_IamTypeName[7:13]:       Folder,
	_IamTypeLowerName[7:13]:  Folder,
	_IamTypeName[13:25]:      Organization,
	_IamTypeLowerName[13:25]: Organization,
	_IamTypeName[25:31]:      GSuite,
	_IamTypeLowerName[25:31]: GSuite,
	_IamTypeName[31:38]:      Service,
	_IamTypeLowerName[31:38]: Service,
}

var _IamTypeNames = []string{
	_IamTypeName[0:7],
	_IamTypeName[7:13],
	_IamTypeName[13:25],
	_IamTypeName[25:31],
	_IamTypeName[31:38],
}

// IamTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func IamTypeString(s string) (IamType, error) {
	if val, ok := _IamTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _IamTypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to IamType values", s)
}

// IamTypeValues returns all values of the enum
func IamTypeValues() []IamType {
	return _IamTypeValues
}

// IamTypeStrings returns a slice of all String values of the enum
func IamTypeStrings() []string {
	strs := make([]string, len(_IamTypeNames))
	copy(strs, _IamTypeNames)
	return strs
}

// IsAIamType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i IamType) IsAIamType() bool {
	for _, v := range _IamTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalGQL implements the graphql.Marshaler interface for IamType
func (i IamType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(i.String()))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface for IamType
func (i *IamType) UnmarshalGQL(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("IamType should be a string, got %T", value)
	}

	var err error
	*i, err = IamTypeString(str)
	return err
}
