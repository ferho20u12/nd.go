package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"framework-back/nucleo-de-diagnostico/foundation"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
)

var (
	ErrAutoFieldsConfig     = errors.New("autofields config bad declaration") // HTTP 500 Internal Server Error
	ErrNoJSONDataFound      = errors.New("no JSON data found in the request") // HTTP 400 Bad Request
	ErrTooManyFiles         = errors.New("too many files uploaded")           // HTTP 400 Bad Request
	ErrFileTooLarge         = errors.New("file is too large")                 // HTTP 400 Bad Request
	ErrInvalidFileExtension = errors.New("file has an invalid extension")     // HTTP 400 Bad Request
)

// NewBind initializes a new Bind instance with auto fields configuration.
func NewBind() *Bind {
	autoFields, ok := foundation.App.Config.Get("bind.autofields", nil).(map[string]AutoFields)
	if !ok {
		color.Redln(ErrAutoFieldsConfig)
		os.Exit(1)
	}
	return &Bind{
		autoFields: autoFields,
	}
}

// Bind is the main structure for data binding.
type Bind struct {
	autoFields map[string]AutoFields
}

// AutoFields represents the configuration for automatic fields.
type AutoFields struct {
	Values map[string]ConfigValue
	Fields map[string]string
	Tags   []ConfigTag
}

// ConfigValue represents a value configuration.
type ConfigValue struct {
	Value       interface{}
	ContextFunc func(ctx *gin.Context) (interface{}, error)
	TypeFunc    func(ctx *gin.Context, objType reflect.Type) (interface{}, error)
}

// ConfigTag represents a tag configuration.
type ConfigTag struct {
	Name  string
	Key   string
	Value string
}

// ConfigJson represents the configuration for JSON data.
type ConfigJson struct {
	Obj        interface{}
	Mode       string
	AutoFields *AutoFields
}

// ConfigUrl represents the configuration for URLs with parameters.
type ConfigUrl struct {
	Uris   interface{}
	Params interface{}
}

// FileData represents the data for each file.
type FileData struct {
	File *multipart.FileHeader
	Size int64
	Ext  string
}

// FileParam represents the parameters for files in the request.
type FileParam struct {
	Param            string
	FilesAllowed     int
	AllowedExtension string
	FilesDatas       []FileData
}

// ConfigFormData represents the configuration for multipart form data.
type ConfigFormData struct {
	MaxFileSize int64        // in bytes, defaults to 20MB if not specified
	FilesParams *[]FileParam // Using a pointer to modify the original object
	ConfigJson  *ConfigJson
	ConfigUrl   *ConfigUrl
}

// Url handles binding of URI and query parameters.
func (b *Bind) Url(ctx *gin.Context, config ConfigUrl) error {
	if config.Uris != nil {
		if err := ctx.ShouldBindUri(config.Uris); err != nil {
			return err
		}
	}
	if config.Params != nil {
		if err := ctx.ShouldBindQuery(config.Params); err != nil {
			return err
		}
	}
	return nil
}

// Json handles binding of a JSON body to a given struct and query parameters and auto fields.
func (b *Bind) Json(ctx *gin.Context, config ConfigJson, configUrl *ConfigUrl) error {
	if err := ctx.ShouldBindJSON(config.Obj); err != nil {
		return err
	}
	if configUrl != nil {
		if err := b.Url(ctx, *configUrl); err != nil {
			return err
		}
	}
	return b.handleAutoFields(ctx, config)
}

// FormData handles binding of multipart form data (including files) to a given struct.
func (b *Bind) FormData(ctx *gin.Context, config ConfigFormData) error {
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 20 * 1024 * 1024 // Default to 20MB
	}
	if config.FilesParams != nil {
		for i := range *config.FilesParams {
			fileParam := &(*config.FilesParams)[i] // Obtain a pointer to modify the original object
			if err := b.handleFiles(ctx, fileParam, config.MaxFileSize); err != nil {
				return err
			}
		}
	}
	if config.ConfigJson != nil {
		form, err := ctx.MultipartForm()
		if err != nil {
			return err
		}
		jsonDataArray, exists := form.Value["dataJSON"]
		if !exists || len(jsonDataArray) != 1 {
			return ErrNoJSONDataFound
		}
		jsonData := jsonDataArray[0]
		if err := json.Unmarshal([]byte(jsonData), config.ConfigJson.Obj); err != nil {
			return err
		}
		if err := b.handleAutoFields(ctx, *config.ConfigJson); err != nil {
			return err
		}
	}
	if config.ConfigUrl != nil {
		if err := b.Url(ctx, *config.ConfigUrl); err != nil {
			return err
		}
	}
	return nil
}

// handleFiles processes file uploads based on the FileParam configuration.
func (b *Bind) handleFiles(ctx *gin.Context, fileParam *FileParam, maxFileSize int64) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File[fileParam.Param]
	if len(files) > fileParam.FilesAllowed {
		return ErrTooManyFiles
	}
	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return fmt.Errorf("%w: %s", ErrFileTooLarge, fileHeader.Filename)
		}
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if fileParam.AllowedExtension != "" && !strings.Contains(fileParam.AllowedExtension, ext) {
			return fmt.Errorf("%w: %s", ErrInvalidFileExtension, fileHeader.Filename)
		}
		fileParam.FilesDatas = append(fileParam.FilesDatas, FileData{
			File: fileHeader,
			Size: fileHeader.Size,
			Ext:  ext,
		})
	}
	return nil
}

// handleAutoFields handles automatic field population based on the mode and auto fields configuration.
func (b *Bind) handleAutoFields(ctx *gin.Context, config ConfigJson) error {
	if config.AutoFields != nil || config.Mode != "" {
		objType := reflect.TypeOf(config.Obj).Elem()
		var autoFields AutoFields
		if config.AutoFields != nil {
			autoFields = *config.AutoFields
		} else {
			autoFields = b.autoFields[config.Mode]
		}
		autoFieldsMap, err := b.GetAutoFields(ctx, objType, autoFields)
		if err != nil {
			return err
		}
		if autoFieldsMap != nil {
			b.fillAutoFields(config.Obj, autoFieldsMap)
		}
	}
	return nil
}

// GetAutoFields retrieves auto fields based on the object type and mode.
func (b *Bind) GetAutoFields(ctx *gin.Context, objType reflect.Type, autoFields AutoFields) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	autofieldsMap := make(map[string]interface{})
	for name, v := range autoFields.Values {
		if v.ContextFunc != nil {
			res, err := v.ContextFunc(ctx)
			if err != nil {
				return nil, err
			}
			values[name] = res
		} else if v.TypeFunc != nil {
			res, err := v.TypeFunc(ctx, objType)
			if err != nil {
				return nil, err
			}
			values[name] = res
		} else {
			values[name] = v.Value
		}
	}
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		for name, value := range autoFields.Fields {
			if strings.Contains(field.Name, name) {
				if val, exists := values[value]; exists {
					autofieldsMap[field.Name] = val
				}
				break
			}
		}
		if _, found := autofieldsMap[field.Name]; !found {
			for _, t := range autoFields.Tags {
				if strings.Contains(field.Tag.Get(t.Name), t.Key) {
					if val, exists := values[t.Value]; exists {
						autofieldsMap[field.Name] = val
					}
					break
				}
			}
		}
	}

	return autofieldsMap, nil
}

func (b *Bind) fillAutoFields(obj interface{}, autoFields map[string]interface{}) {
	objValue := reflect.ValueOf(obj).Elem()
	for field, value := range autoFields {
		f := objValue.FieldByName(field)
		if f.IsValid() && f.CanSet() {
			f.Set(reflect.ValueOf(value))
		}
	}
}
