package orm

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func SearchWithCount(where Where, pager Pager, entityName string, references ...string) (results []interface{}, totalRows int) {
	return search(where, pager, true, entityName, references)
}

func Search(where Where, pager Pager, entityName string, references ...string) []interface{} {
	results, _ := search(where, pager, false, entityName, references)
	return results
}

func SearchOne(where Where, pager Pager, entityName string, references ...string) interface{} {
	results, _ := search(where, pager, false, entityName, references)
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

func SearchIdsWithCount(where Where, pager Pager, entityName string) (results []uint64, totalRows int) {
	return searchIds(where, pager, true, entityName)
}

func SearchIds(where Where, pager Pager, entityName string) []uint64 {
	results, _ := searchIds(where, pager, false, entityName)
	return results
}

func search(where Where, pager Pager, withCount bool, entityName string, references []string) ([]interface{}, int) {
	entityType := getEntityType(entityName)
	schema := GetTableSchema(entityName)

	var fieldsList = buildFieldList(entityType, "")
	query := fmt.Sprintf("SELECT CONCAT_WS('|', %s) FROM `%s` WHERE %s %s", fieldsList, schema.TableName, where, pager.String())
	results, err := schema.GetMysqlDB().Query(query, where.GetParameters()...)
	if err != nil {
		panic(err.Error())
	}
	result := make([]interface{}, 0, pager.GetPageSize())
	for results.Next() {
		var row string
		err = results.Scan(&row)
		if err != nil {
			panic(err.Error())
		}
		entity := createEntityFromDBRow(row, entityType)
		result = append(result, entity)
	}
	totalRows := getTotalRows(withCount, pager, where, schema, len(result))
	warmUpReferences(schema, result, references)
	return result, totalRows
}

func searchIds(where Where, pager Pager, withCount bool, entityName string) ([]uint64, int) {
	schema := GetTableSchema(entityName)
	query := fmt.Sprintf("SELECT `Id` FROM `%s` WHERE %s %s", schema.TableName, where, pager.String())
	results, err := schema.GetMysqlDB().Query(query, where.GetParameters()...)
	if err != nil {
		panic(err.Error())
	}
	result := make([]uint64, 0, pager.GetPageSize())
	for results.Next() {
		var row uint64
		err = results.Scan(&row)
		if err != nil {
			panic(err.Error())
		}
		result = append(result, row)
	}
	totalRows := getTotalRows(withCount, pager, where, schema, len(result))
	return result, totalRows
}

func getTotalRows(withCount bool, pager Pager, where Where, schema *TableSchema, foundRows int) int {
	totalRows := 0
	if withCount {
		totalRows = foundRows
		if totalRows == pager.GetPageSize() {
			query := fmt.Sprintf("SELECT count(1) FROM `%s` WHERE %s", schema.TableName, where)
			var foundTotal string
			err := schema.GetMysqlDB().QueryRow(query, where.GetParameters()...).Scan(&foundTotal)
			if err != nil {
				panic(err.Error())
			}
			totalRows, _ = strconv.Atoi(foundTotal)
		} else {
			totalRows += (pager.GetCurrentPage() - 1) * pager.GetPageSize()
		}
	}
	return totalRows
}

func createEntityFromDBRow(row string, entityType reflect.Type) interface{} {
	data := strings.Split(row, "|")
	value := reflect.New(entityType).Elem()

	fillStruct(0, data, entityType, value, "")
	orm := value.Field(0).Interface().(ORM)
	orm.DBData = make(map[string]interface{})
	orm.DBData["Id"] = data[0]
	value.Field(0).Set(reflect.ValueOf(orm))
	entity := value.Interface()

	_, bind := IsDirty(entity)
	cc := reflect.ValueOf(entity).Field(0).Interface().(ORM)
	for key, value := range bind {
		cc.DBData[key] = value
	}
	return entity
}

func fillStruct(index uint16, data []string, t reflect.Type, value reflect.Value, prefix string) uint16 {

	bind := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {

		if index == 0 && i == 0 {
			continue
		}

		field := value.Field(i)
		name := prefix + t.Field(i).Name

		fieldType := field.Type().String()
		switch fieldType {
		case "uint":
			bind[name] = data[index]
			integer, _ := strconv.ParseUint(data[index], 10, 32)
			field.SetUint(integer)
		case "uint8":
			integer, _ := strconv.ParseUint(data[index], 10, 8)
			field.SetUint(integer)
		case "uint16":
			integer, _ := strconv.ParseUint(data[index], 10, 16)
			field.SetUint(integer)
		case "uint32":
			integer, _ := strconv.ParseUint(data[index], 10, 32)
			field.SetUint(integer)
		case "uint64":
			integer, _ := strconv.ParseUint(data[index], 10, 64)
			field.SetUint(integer)
		case "int":
			integer, _ := strconv.ParseInt(data[index], 10, 32)
			field.SetInt(integer)
		case "int8":
			integer, _ := strconv.ParseInt(data[index], 10, 8)
			field.SetInt(integer)
		case "int16":
			integer, _ := strconv.ParseInt(data[index], 10, 16)
			field.SetInt(integer)
		case "int32":
			integer, _ := strconv.ParseInt(data[index], 10, 32)
			field.SetInt(integer)
		case "int64":
			integer, _ := strconv.ParseInt(data[index], 10, 64)
			field.SetInt(integer)
		case "string":
			field.SetString(data[index])
		case "[]string":
			if data[index] != "" {
				var values = strings.Split(data[index], ",")
				var length = len(values)
				slice := reflect.MakeSlice(field.Type(), length, length)
				for key, value := range values {
					slice.Index(key).SetString(value)
				}
				field.Set(slice)
			}
		case "[]uint64":
			if data[index] != "" {
				var values = strings.Split(data[index], " ")
				var length = len(values)
				slice := reflect.MakeSlice(field.Type(), length, length)
				for key, value := range values {
					integer, _ := strconv.ParseUint(value, 10, 64)
					slice.Index(key).SetUint(integer)
				}
				field.Set(slice)
			}

		case "bool":
			val := false
			if data[index] == "1" {
				val = true
			}
			field.SetBool(val)
		case "float32":
			float, _ := strconv.ParseFloat(data[index], 32)
			field.SetFloat(float)
		case "float64":
			float, _ := strconv.ParseFloat(data[index], 64)
			field.SetFloat(float)
		case "time.Time":
			layout := "2006-01-02"
			if len(data[index]) == 19 {
				layout += " 15:04:05"
			}
			value, _ := time.Parse(layout, data[index])
			field.Set(reflect.ValueOf(value))
		case "interface {}":
			if data[index] != "" {
				var f interface{}
				err := json.Unmarshal([]byte(data[index]), &f)
				if err != nil {
					panic(fmt.Errorf("invalid json: %s", data[index]))
				}
				field.Set(reflect.ValueOf(f))
			}
		default:
			if field.Kind().String() == "struct" {
				newVal := reflect.New(field.Type())
				value := newVal.Elem()
				index = fillStruct(index, data, field.Type(), value, name)
				field.Set(value)
				continue
			}
			panic(fmt.Errorf("unsoported field type: %s", field.Type().String()))
		}
		index++
	}
	return index
}

func buildFieldList(t reflect.Type, prefix string) string {
	fieldsList := ""
	for i := 0; i < t.NumField(); i++ {
		var columnNameRaw string
		field := t.Field(i)
		if prefix == "" && (field.Name == "Id" || field.Name == "Orm") {
			continue
		}
		switch field.Type.String() {
		case "string", "[]string", "interface {}":
			columnNameRaw = prefix + t.Field(i).Name
			fieldsList += fmt.Sprintf(",IFNULL(`%s`,'')", columnNameRaw)
		default:
			if field.Type.Kind().String() == "struct" && field.Type.String() != "time.Time" {
				fieldsList += buildFieldList(field.Type, field.Name)
			} else {
				columnNameRaw = prefix + t.Field(i).Name
				fieldsList += fmt.Sprintf(",`%s`", columnNameRaw)
			}
		}
	}
	if prefix == "" {
		fieldsList = "`Id`" + fieldsList
	}
	return fieldsList
}
