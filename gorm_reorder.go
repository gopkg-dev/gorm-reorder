package gorm_reorder

import (
	"log"
	"reflect"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type dependency struct {
	*gorm.Statement
	Depends []*schema.Schema
}

type reorder struct {
	db                            *gorm.DB
	autoAdd                       bool
	modelNames, orderedModelNames []string
	orderedModelNamesMap          map[string]bool
	parsedSchemas                 map[*schema.Schema]bool
	valuesMap                     map[string]dependency
	models                        []interface{}
	schemas                       []*schema.Schema
}

type Config struct {
	AutoAdd       bool
	TablePrefix   string // 表名前缀，`User` 的表名应该是 `t_users`
	SingularTable bool   // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
}

func NewReorder(cfg Config) *reorder {
	dsn := "file::memory:?cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix,
			SingularTable: cfg.SingularTable,
		}})
	if err != nil {
		return nil
	}
	return &reorder{
		db:                   db,
		autoAdd:              cfg.AutoAdd,
		orderedModelNamesMap: map[string]bool{},
		parsedSchemas:        map[*schema.Schema]bool{},
		valuesMap:            map[string]dependency{},
	}
}

func (r *reorder) AddModel(dst []interface{}) *reorder {
	if len(dst) > 0 {
		for _, m := range dst {
			r.models = append(r.models, m)
		}
	}
	return r
}

func (r *reorder) parseDependence(value interface{}, addToList bool) {
	dep := dependency{
		Statement: &gorm.Statement{DB: r.db, Dest: value},
	}
	beDependedOn := map[*schema.Schema]bool{}
	if err := dep.Parse(value); err != nil {
		log.Fatalf("failed to parse value %#v, got error %v", value, err)
	}
	if _, ok := r.parsedSchemas[dep.Statement.Schema]; ok {
		return
	}
	r.parsedSchemas[dep.Statement.Schema] = true

	for _, rel := range dep.Schema.Relationships.Relations {
		if c := rel.ParseConstraint(); c != nil && c.Schema == dep.Statement.Schema && c.Schema != c.ReferenceSchema {
			dep.Depends = append(dep.Depends, c.ReferenceSchema)
		}

		if rel.Type == schema.HasOne || rel.Type == schema.HasMany {
			beDependedOn[rel.FieldSchema] = true
		}

		if rel.JoinTable != nil {
			// append join value
			defer func(rel *schema.Relationship, joinValue interface{}) {
				if !beDependedOn[rel.FieldSchema] {
					dep.Depends = append(dep.Depends, rel.FieldSchema)
				} else {
					fieldValue := reflect.New(rel.FieldSchema.ModelType).Interface()
					r.parseDependence(fieldValue, r.autoAdd)
				}
				r.parseDependence(joinValue, r.autoAdd)
			}(rel, reflect.New(rel.JoinTable.ModelType).Interface())
		}
	}

	r.valuesMap[dep.Schema.Table] = dep

	if addToList {
		r.modelNames = append(r.modelNames, dep.Schema.Table)
	}
}

func (r *reorder) insertIntoOrderedList(name string) {
	if _, ok := r.orderedModelNamesMap[name]; ok {
		return // avoid loop
	}
	r.orderedModelNamesMap[name] = true

	if r.autoAdd {
		dep := r.valuesMap[name]
		for _, d := range dep.Depends {
			if _, ok := r.valuesMap[d.Table]; ok {
				r.insertIntoOrderedList(d.Table)
			} else {
				r.parseDependence(reflect.New(d.ModelType).Interface(), r.autoAdd)
				r.insertIntoOrderedList(d.Table)
			}
		}
	}

	r.orderedModelNames = append(r.orderedModelNames, name)
}

func (r *reorder) Parser() *reorder {

	results := make([]interface{}, 0)

	for _, value := range r.models {
		if v, ok := value.(string); ok {
			results = append(results, v)
		} else {
			r.parseDependence(value, true)
		}
	}

	for _, name := range r.modelNames {
		r.insertIntoOrderedList(name)
	}

	for _, name := range r.orderedModelNames {
		results = append(results, r.valuesMap[name].Statement.Dest)
	}

	for s := range r.parsedSchemas {
		r.schemas = append(r.schemas, s)
	}

	return r
}

func (r *reorder) GetSchemas() []*schema.Schema {
	return r.schemas
}
