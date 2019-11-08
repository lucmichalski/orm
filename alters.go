package orm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func GetAlters() (safeAlters []string, unsafeAlters []string) {

	tablesInDB := make(map[string]map[string]bool)
	tablesInEntities := make(map[string]map[string]bool)

	for _, poolName := range mysqlPoolCodes {
		tablesInDB[poolName] = make(map[string]bool)
		results, err := GetMysqlDB(poolName).Query("SHOW TABLES")
		if err != nil {
			panic(err.Error())
		}
		for results.Next() {
			var row string
			err = results.Scan(&row)
			if err != nil {
				panic(err.Error())
			}
			tablesInDB[poolName][row] = true
		}
		tablesInEntities[poolName] = make(map[string]bool)
	}
	for name := range entities {
		tableSchema := GetTableSchema(name)
		tablesInEntities[tableSchema.MysqlPoolName][tableSchema.TableName] = true
		has, safeAlter, unsafeAlter := tableSchema.GetSchemaChanges()
		if !has {
			continue
		}
		if safeAlter != "" {
			safeAlters = append(safeAlters, safeAlter)
		} else if unsafeAlter != "" {
			unsafeAlters = append(unsafeAlters, unsafeAlter)
		}
	}

	for poolName, tables := range tablesInDB {
		for tableName := range tables {
			_, has := tablesInEntities[poolName][tableName]
			if !has {
				dropSql := fmt.Sprintf("DROP TABLE `%s`;", tableName)
				if isTableEmptyInPool(poolName, tableName) {
					safeAlters = append(safeAlters, dropSql)
				} else {
					unsafeAlters = append(unsafeAlters, dropSql)
				}
			}
		}
	}
	return
}

func isTableEmptyInPool(poolName string, tableName string) bool {
	var lastId uint64
	err := GetMysqlDB(poolName).QueryRow(fmt.Sprintf("SELECT `Id` FROM `%s` LIMIT 1", tableName)).Scan(&lastId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return true
		}
		panic(err.Error())
	}
	return false
}