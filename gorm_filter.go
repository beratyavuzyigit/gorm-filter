package gormfilter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormFilter struct {
	tx *gorm.DB

	joinList     []string
	tableList    []string
	joinedTables []string
}

func (db *DB) Query(conds map[string]string) *DB {
	gormFilter := &GormFilter{tx: db.DB}
	gormFilter.PrepareQuery(conds)
	return db
}

func (db *DB) Limit(limit int) *DB {
	if limit <= 0 {
		return db
	}
	db.DB = db.DB.Limit(limit)
	return db
}

func (db *DB) Offset(offset int) *DB {
	if offset <= 0 {
		return db
	}
	db.DB = db.DB.Offset(offset)
	return db
}

func (q *GormFilter) PrepareQuery(conds map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			q.tx.Statement.AddError(fmt.Errorf("%v", r))
		}
	}()

	// Reset the query
	q.tableList = make([]string, 0)
	q.joinedTables = make([]string, 0)

	// Get the main model
	mainModel := q.tx.Statement.Model

	// Loop through the conditions
	for key, value := range conds {
		q.tableList = append(q.tableList, q.getTableName(mainModel)) // Add the main model to the table list

		fields := strings.Split(key, SEPERATOR) // Split the key by the seperator
		if len(fields) == 1 {
			tableName := q.getTableName(mainModel)
			field := q.getFieldByJsonTagName(mainModel, key)
			column_type := q.getFieldTypeByField(field)
			err := q.appendWhereList(key, column_type, value, tableName, "contains")
			if err != nil {
				q.tx.Statement.AddError(err)
			}
		} else {
			fieldList := make([]string, 0)
			model := mainModel
			var columnName string
			var columnType string
			var searchType string

			for _, field := range fields {
				if inArray(SearchTypes, field) {
					searchType = field
				} else {
					fieldList = append(fieldList, field)
					modelField := q.getFieldByJsonTagName(model, field)
					fkModel := q.getFKModelByModelField(model, modelField)
					columnType = q.getFieldTypeByField(modelField)
					if fkModel == nil {
						columnName = field
					} else {
						model = fkModel
						model_name := reflect.TypeOf(fkModel).Name()
						table_name := q.tx.NamingStrategy.TableName(model_name)
						q.appendJoinList(table_name, q.getColumnNameByField(modelField))
						q.tableList = append(q.tableList, table_name)
					}
				}
			}
			if columnName == "" {
				columnType = DEFAULT_PK_TYPE
				q.joinList = q.joinList[:len(q.joinList)-1]
				q.tableList = q.tableList[:len(q.tableList)-1]
			}
			lastTableName := q.tableList[len(q.tableList)-1]
			columnName = fieldList[len(fieldList)-1]
			if searchType == "" {
				if columnName == "" {
					columnName = DEFAULT_PK
				}
				err := q.appendWhereList(columnName, columnType, value, lastTableName, "contains")
				if err != nil {
					q.tx.Statement.AddError(err)
				}
				continue
			}
			err := q.appendWhereList(columnName, columnType, value, lastTableName, SearchType(searchType))
			if err != nil {
				q.tx.Statement.AddError(err)
			}
		}
	}
	for _, joinQuery := range q.joinList {
		q.tx.Joins(joinQuery)
	}
}

func (q *GormFilter) appendWhereList(field, fieldType, value, tableName string, searchType SearchType) error {
	var fieldName string
	if tableName == "" {
		fieldName = field
	} else {
		fieldName = tableName + "." + field
	}

	if fieldType == "bool" {
		searchType = "e"
	}

	if fieldType == "bool" && value != "true" && value != "false" {
		return ErrBoolTypeNeedsTrueOrFalse
	}

	if searchType == "contains" {
		if fieldType == "int" {
			q.tx.Where(fieldName+" = ?", value)
		} else if _, err := uuid.Parse(value); err == nil {
			q.tx.Where(fieldName+"::text LIKE ?", "%"+value+"%")
		} else {
			q.tx.Where(fieldName+" LIKE ?", "%"+value+"%")
		}
	} else if searchType == GREATER_THAN {
		q.tx.Where(fieldName+" > ?", value)
	} else if searchType == GREATER_THAN_OR_EQUAL {
		q.tx.Where(fieldName+" >= ?", value)
	} else if searchType == LESS_THAN {
		q.tx.Where(fieldName+" < ?", value)
	} else if searchType == LESS_THAN_OR_EQUAL {
		q.tx.Where(fieldName+" <= ?", value)
	} else if searchType == EQUAL {
		if _, err := uuid.Parse(value); err == nil {
			q.tx.Where(fieldName+"::text = ?", value)
		} else {
			if fieldType == "uuid" {
				q.tx.Where(fieldName + " IS NULL")
			} else {
				q.tx.Where(fieldName+" = ?", value)
			}
		}
	} else if searchType == NOT_EQUAL {
		if _, err := uuid.Parse(value); err == nil {
			q.tx.Where(fieldName+"::text != ?", value)
		} else {
			if fieldType == "uuid" {
				q.tx.Where(fieldName + " IS NOT NULL")
			} else {
				q.tx.Where(fieldName+" != ?", value)
			}
		}
	} else if searchType == IN {
		q.tx.Where(fieldName+" IN ?", strings.Split(value, ";"))
	} else {
		return ErrUndefinedSearchType
	}

	return nil
}

func (q *GormFilter) appendJoinList(tableName, fkFieldName string) {
	if inArray(q.joinedTables, tableName) {
		return
	}
	fkTableName := q.tableList[len(q.tableList)-1]
	q.joinedTables = append(q.joinedTables, tableName)
	var joinQuery string = "INNER JOIN " + tableName + " ON " + tableName + ".id = " + fkTableName + "." + fkFieldName
	q.joinList = append(q.joinList, joinQuery)
}

func (q *GormFilter) getTableName(model interface{}) string {
	return q.tx.NamingStrategy.TableName(reflect.TypeOf(model).Name())
}

func (q *GormFilter) getFieldByJsonTagName(model interface{}, tag_name string) reflect.StructField {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = reflect.Indirect(modelValue)
	}
	modelType := modelValue.Type()
	numFields := modelType.NumField()
	var field_value reflect.StructField
	for i := 0; i < numFields; i++ {
		field := modelType.Field(i)
		json_tag := field.Tag.Get("json")
		if json_tag == tag_name {
			field_value = field
			break
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedType := field.Type
			numEmbeddedFields := embeddedType.NumField()
			for j := 0; j < numEmbeddedFields; j++ {
				embeddedField := embeddedType.Field(j)
				embeddedJsonTag := embeddedField.Tag.Get("json")
				if embeddedJsonTag == tag_name {
					field_value = embeddedField
					return field_value
				}
			}
		}
	}
	return field_value
}

func (q *GormFilter) getColumnNameByField(field reflect.StructField) string {
	gorm_tag := field.Tag.Get("gorm")
	tag_parts := strings.Split(gorm_tag, ";")
	for _, tagPart := range tag_parts {
		if strings.HasPrefix(strings.TrimSpace(tagPart), "column:") {
			return strings.Split(tagPart, ":")[1]
		}
	}
	return ""
}

func (q *GormFilter) getFKModelByModelField(searchingModel interface{}, modelField reflect.StructField) interface{} {
	field_name := modelField.Name
	searching_model_type := reflect.TypeOf(searchingModel)
	num_fields := searching_model_type.NumField()
	var found_model interface{}
	for i := 0; i < num_fields; i++ {
		field := searching_model_type.Field(i)
		gorm_tag := field.Tag.Get("gorm")
		tag_parts := strings.Split(gorm_tag, ";")
		for _, tagPart := range tag_parts {
			if strings.HasPrefix(strings.TrimSpace(tagPart), "foreignKey:") {
				if strings.Split(tagPart, ":")[1] == field_name {
					found_model = reflect.New(field.Type).Elem().Interface()
					break
				}
			}
		}
	}
	return found_model
}

func (q *GormFilter) getFieldTypeByField(field reflect.StructField) string {
	defer func() {
		if r := recover(); r != nil {
			panic(ErrFieldAppearsToBeUndefined)
		}
	}()
	fieldType := field.Type.String()
	if strings.Contains(fieldType, "uuid.UUID") {
		return "uuid"
	} else if strings.Contains(fieldType, "string") {
		return "string"
	} else if strings.Contains(fieldType, "int") {
		return "int"
	} else if strings.Contains(fieldType, "bool") {
		return "bool"
	} else if strings.Contains(fieldType, "time.Time") {
		return "time"
	}
	return ""
}

func inArray(array []string, val string) bool {
	for _, item := range array {
		if item == val {
			return true
		}
	}
	return false
}
