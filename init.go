package orm

import (
	"reflect"
)

func initIfNeeded(engine *Engine, entity Entity) *ORM {
	orm := entity.getORM()
	if orm.dBData == nil {
		value := reflect.ValueOf(entity)
		elem := value.Elem()
		t := elem.Type()
		tableSchema := getTableSchema(engine.registry, t)
		if tableSchema == nil {
			panic(EntityNotRegisteredError{Name: t.String()})
		}
		orm.engine = engine
		orm.tableSchema = tableSchema
		orm.dBData = make(map[string]interface{}, len(tableSchema.columnNames))
		orm.attributes = &entityAttributes{nil, false, false, value, elem, elem.Field(1), nil}
		defaultInterface, is := entity.(DefaultValuesInterface)
		if is {
			defaultInterface.SetDefaults()
		}
	}
	return orm
}
