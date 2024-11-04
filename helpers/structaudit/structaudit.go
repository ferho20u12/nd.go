package structaudit

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Error messages
var (
	ErrNotPointer               = errors.New("object must be a pointer")                                 // HTTP 400 Bad Request
	ErrInvalidStructure         = errors.New("object must be a valid structure")                         // HTTP 400 Bad Request
	ErrFieldNotExist            = errors.New("field specified in the structure does not exist")          // HTTP 400 Bad Request
	ErrMarshalJSON              = errors.New("failed to marshal data to JSON")                           // HTTP 500 Internal Server Error
	ErrUnmarshalJSON            = errors.New("failed to unmarshal JSON data")                            // HTTP 500 Internal Server Error
	ErrScanValue                = errors.New("failed to scan value")                                     // HTTP 500 Internal Server Error
	ErrUnmarshalData            = errors.New("failed to unmarshal data")                                 // HTTP 500 Internal Server Error
	ErrObjectNotStruct          = errors.New("object is not a structure")                                // HTTP 400 Bad Request
	ErrFieldNotFound            = errors.New("field not found in the structure")                         // HTTP 400 Bad Request
	ErrObjectNotPointer         = errors.New("object is not a pointer")                                  // HTTP 400 Bad Request
	ErrObjectNotSlice           = errors.New("object is not a slice")                                    // HTTP 400 Bad Request
	ErrFieldNotExistInElement   = errors.New("field does not exist in one of the elements of the slice") // HTTP 400 Bad Request
	ErrObjectNotStructForTag    = errors.New("object is not a struct")                                   // HTTP 400 Bad Request
	ErrMethodNotFound           = errors.New("method not found in the structure")                        // HTTP 500 Internal Server Error
	ErrMethodHasParameters      = errors.New("method should not have parameters")                        // HTTP 400 Bad Request
	ErrMethodInvalidReturnCount = errors.New("method should return a single value")                      // HTTP 500 Internal Server Error
	ErrMethodInvalidReturnType  = errors.New("method should return a string value")                      // HTTP 500 Internal Server Error
)

// FieldInfo holds information about a field in a structure
type FieldInfo struct {
	Name    string
	TagJson string
	Value   any
	Type    reflect.Type
}

// NormalizePointerType ensures that the given object is a pointer and returns its type.
func NormalizePointerType(obj any) (reflect.Type, error) {
	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		return nil, ErrNotPointer
	}
	typ = typ.Elem()
	if typ.Kind() == reflect.Slice {
		typ = reflect.TypeOf(obj).Elem().Elem()
	} else if typ.Kind() == reflect.Struct {
		typ = reflect.TypeOf(obj).Elem()
	} else {
		return nil, ErrInvalidStructure
	}
	return typ, nil
}

// GetObjectKind returns the kind of the object.
func GetObjectKind(obj any) reflect.Kind {
	typ := reflect.TypeOf(obj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.Kind()
}

// LocateFieldType locates the type of a field in a structure by its path.
func LocateFieldType(objType reflect.Type, fieldPath string, maxDepth int) (reflect.Type, error) {
	parts := strings.Split(fieldPath, ".")
	if (maxDepth == -1 || maxDepth > 0) && len(parts) > 1 {
		maxDepth--
	}
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if field.Name == parts[0] {
			if len(parts) > 1 && ((maxDepth == -1 || maxDepth > 0) && field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
				fieldType := field.Type.Elem()
				return LocateFieldType(fieldType, strings.Join(parts[1:], "."), maxDepth)
			}
			return field.Type, nil
		}
	}
	return nil, ErrFieldNotExist
}

func FindFieldInfoByTag(structType reflect.Type, tagName, tagValue string) (*FieldInfo, error) {
	var fieldInfo *FieldInfo
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get(tagName)
		if strings.Contains(tag, tagValue) {
			if fieldInfo == nil {
				fieldType := field.Type
				fieldValue := reflect.New(fieldType)
				fieldInfo = &FieldInfo{
					Name:  field.Name,
					Type:  fieldType,
					Value: fieldValue,
				}
			} else {
				return nil, fmt.Errorf("multiple fields found with tag '%s' containing '%s' for type %s", tagName, tagValue, structType.String())
			}
		}
	}

	if fieldInfo == nil {
		return nil, fmt.Errorf("no field found with tag '%s' containing '%s' for type %s", tagName, tagValue, structType.String())
	}

	return fieldInfo, nil
}

// ValidateFieldData validates and sets the value of a field.
func ValidateFieldData(fieldInfo *FieldInfo, text string) error {
	var value interface{}
	if fieldInfo.Type.Kind() == reflect.Ptr {
		value = reflect.New(fieldInfo.Type.Elem()).Interface()
	} else {
		value = reflect.New(fieldInfo.Type).Interface()
	}
	wrappedData := map[string]interface{}{
		"Value": text,
	}
	dataBytes, err := json.Marshal(wrappedData)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalJSON, err)
	}
	type Wrapper struct {
		Value interface{} `json:"Value"`
	}
	var wrapper Wrapper
	if err := json.Unmarshal(dataBytes, &wrapper); err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshalJSON, err)
	}
	if err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, wrapper.Value)), value); err != nil {
		if scanner, ok := value.(interface{ Scan(interface{}) error }); ok {
			if err := scanner.Scan(wrapper.Value); err != nil {
				return fmt.Errorf("%w: %v", ErrScanValue, err)
			}
		} else {
			return fmt.Errorf("%w: %v", ErrUnmarshalData, err)
		}
	}
	if fieldInfo.Type.Kind() == reflect.Ptr {
		fieldInfo.Value = reflect.ValueOf(value).Elem().Interface()
	} else {
		fieldInfo.Value = value
	}
	return nil
}

// FindFieldInfoByName searches for a field by its name.
func FindFieldInfoByName(structType reflect.Type, fieldName string) (*FieldInfo, error) {
	var fieldInfo *FieldInfo
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.EqualFold(field.Name, fieldName) {
			if fieldInfo == nil {
				fieldType := field.Type
				var fieldValue reflect.Value
				if fieldType.Kind() == reflect.Ptr {
					fieldValue = reflect.New(fieldType.Elem())
				} else {
					fieldValue = reflect.Zero(fieldType)
				}

				fieldInfo = &FieldInfo{
					Name:  field.Name,
					Type:  fieldType,
					Value: fieldValue,
				}
			} else {
				return nil, fmt.Errorf("multiple fields found with name '%s' for type %s", fieldName, structType.String())
			}
		}
	}
	if fieldInfo == nil {
		return nil, fmt.Errorf("no field found with name '%s' for type %s", fieldName, structType.String())
	}

	return fieldInfo, nil
}

// RetrieveFieldData retrieves the value of a field from a structure.
func RetrieveFieldData(obj interface{}, fieldName string) (interface{}, error) {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}
	if objValue.Kind() != reflect.Struct {
		return nil, ErrObjectNotStruct
	}
	fieldValue := objValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return nil, ErrFieldNotFound
	}
	return fieldValue.Interface(), nil
}

// ExtractCollectionFromField extracts a collection of field values from a slice of structures.
func ExtractCollectionFromField(obj interface{}, fieldName string) ([]interface{}, error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return nil, ErrObjectNotPointer
	}
	objType = objType.Elem()
	if objType.Kind() != reflect.Slice {
		return nil, ErrObjectNotSlice
	}
	values := make([]interface{}, 0)
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}
	for i := 0; i < objValue.Len(); i++ {
		fieldValue := objValue.Index(i).FieldByName(fieldName)
		if !fieldValue.IsValid() {
			return nil, ErrFieldNotExistInElement
		}
		values = append(values, fieldValue.Interface())
	}

	return values, nil
}

// ExtractFieldsByTag extracts fields by tag name and key.
func ExtractFieldsByTag(objType reflect.Type, tagName string, tagKey string) ([]interface{}, error) {
	if objType.Kind() != reflect.Struct {
		return nil, ErrObjectNotStructForTag
	}
	values := make([]interface{}, 0)
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		tag := field.Tag.Get(tagName)
		if strings.Contains(tag, tagKey) {
			values = append(values, field.Name)
		}
	}
	return values, nil
}

// RetrieveFunctionResult retrieves the result of a method from a type.
func RetrieveFunctionResult(objType reflect.Type, functionName string) (any, error) {
	instance := reflect.New(objType).Interface()
	method := reflect.ValueOf(instance).MethodByName(functionName)
	if !method.IsValid() {
		return nil, ErrMethodNotFound
	}
	if method.Type().NumIn() != 0 {
		return nil, ErrMethodHasParameters
	}
	if method.Type().NumOut() != 1 {
		return nil, ErrMethodInvalidReturnCount
	}
	result := method.Call(nil)[0]
	if result.Kind() != reflect.String {
		return nil, ErrMethodInvalidReturnType
	}
	return result.Interface(), nil
}

// PopulateObjectFields sets the values of fields in an object.
func PopulateObjectFields(obj interface{}, fields map[string]interface{}) {
	objValue := reflect.ValueOf(obj).Elem()
	objType := objValue.Type()
	if objType.Kind() == reflect.Slice {
		for i := 0; i < objValue.Len(); i++ {
			elem := objValue.Index(i)
			elemType := elem.Type()
			if elemType.Kind() == reflect.Struct {
				FillStructFields(elem, fields)
			}
		}
	} else if objType.Kind() == reflect.Struct {
		FillStructFields(objValue, fields)
	}
}

// FillStructFields sets the fields of a structure with values.
func FillStructFields(objValue reflect.Value, fields map[string]interface{}) {
	for fieldName, fieldValue := range fields {
		field := objValue.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}
		if field.CanSet() {
			fieldType := field.Type()
			value := reflect.ValueOf(fieldValue)
			if value.Type().ConvertibleTo(fieldType) {
				convertedValue := value.Convert(fieldType)
				field.Set(convertedValue)
			}
		}
	}
}
