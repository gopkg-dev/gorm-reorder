package gorm_reorder

import (
	"encoding/json"
	"reflect"

	"gorm.io/gorm/schema"
)

type Schema struct {
	Name      string   `json:"name"`       //结构体名称
	TableName string   `json:"table_name"` //数据库表名称
	Fields    []*Field `json:"fields"`     //字段
}

// Field represents an ent.Field that was loaded from a complied user package.
type Field struct {
	Name            string            `json:"name"`              //字段名称
	DBName          string            `json:"db_name"`           //数据库字段名称
	DataType        string            `json:"data_type"`         //字段类型名称
	FullTypeName    string            `json:"full_type_name"`    //
	Size            int               `json:"size"`              //大小
	Unique          bool              `json:"unique"`            //是否唯一字段
	Comment         string            `json:"comment"`           //字段注释
	NotNull         bool              `json:"not_null"`          //是否 NOT NULL
	IsPtr           bool              `json:"is_ptr"`            //是否为指针
	HasDefaultValue bool              `json:"has_default_value"` //是否有默认值
	DefaultValue    interface{}       `json:"default_value"`     //默认值
	PrimaryKey      bool              `json:"primary_key"`       //是否主键
	AutoIncrement   bool              `json:"auto_increment"`    //是否自增
	Tags            map[string]string `json:"tags"`              //Tags
}

func MarshalSchema(s []*schema.Schema) (b []byte, err error) {
	var schemas []Schema
	for _, s := range s {
		entSchema := Schema{
			Name:      s.Name,
			TableName: s.Table,
			Fields:    []*Field{},
		}
		for _, field := range s.Fields {
			if field.IgnoreMigration || field.GORMDataType == "" {
				continue
			}
			entSchema.Fields = append(entSchema.Fields, &Field{
				Name:            field.Name,
				DBName:          field.DBName,
				DataType:        string(field.DataType),
				FullTypeName:    field.FieldType.String(),
				IsPtr:           field.FieldType.Kind() == reflect.Ptr,
				Size:            field.Size,
				Unique:          field.Unique,
				Comment:         field.Comment,
				NotNull:         field.NotNull,
				HasDefaultValue: field.HasDefaultValue,
				DefaultValue:    field.DefaultValue,
				PrimaryKey:      field.PrimaryKey,
				AutoIncrement:   field.AutoIncrement,
				Tags:            field.TagSettings,
			})
		}
		schemas = append(schemas, entSchema)
	}
	return json.MarshalIndent(schemas, "", "\t")
}

func UnmarshalSchema(buf []byte) ([]Schema, error) {
	var schemas []Schema
	if err := json.Unmarshal(buf, &schemas); err != nil {
		return nil, err
	}
	return schemas, nil
}
