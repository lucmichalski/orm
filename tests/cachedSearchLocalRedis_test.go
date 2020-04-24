package tests

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type TestEntityIndexTestLocalRedis struct {
	orm.ORM      `orm:"localCache;redisCache"`
	ID           uint
	Name         string `orm:"length=100;index=FirstIndex"`
	Age          uint16
	Ignore       uint16           `orm:"ignore"`
	IndexAge     *orm.CachedQuery `query:":Age = ? ORDER BY :ID"`
	IndexAll     *orm.CachedQuery `query:""`
	IndexName    *orm.CachedQuery `queryOne:":Name = ?"`
	ReferenceOne *TestEntityIndexTestLocalRedisRef
}

type TestEntityIndexTestLocalRedisRef struct {
	orm.ORM
	ID   uint
	Name string
}

func TestCachedSearchLocalRedis(t *testing.T) {
	var entity *TestEntityIndexTestLocalRedis
	var entityRef *TestEntityIndexTestLocalRedisRef
	engine := PrepareTables(t, &orm.Registry{}, entityRef, entity)

	for i := 1; i <= 5; i++ {
		e := &TestEntityIndexTestLocalRedisRef{Name: "Name " + strconv.Itoa(i)}
		engine.RegisterEntity(e)
		err := e.Flush()
		assert.Nil(t, err)
	}

	var entities = make([]interface{}, 10)
	for i := 1; i <= 5; i++ {
		e := &TestEntityIndexTestLocalRedis{Name: "Name " + strconv.Itoa(i), Age: uint16(10)}
		engine.RegisterEntity(e)
		e.ReferenceOne = &TestEntityIndexTestLocalRedisRef{ID: uint(i)}
		entities[i-1] = e
		err := e.Flush()
		assert.Nil(t, err)
	}
	for i := 6; i <= 10; i++ {
		e := &TestEntityIndexTestLocalRedis{Name: "Name " + strconv.Itoa(i), Age: uint16(18)}
		engine.RegisterEntity(e)
		entities[i-1] = e
		err := e.Flush()
		assert.Nil(t, err)
	}
	pager := &orm.Pager{CurrentPage: 1, PageSize: 100}
	var rows []*TestEntityIndexTestLocalRedis
	totalRows, err := engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.True(t, rows[0].Loaded())

	assert.Len(t, rows, 5)
	assert.Equal(t, uint(1), rows[0].ReferenceOne.ID)
	assert.Equal(t, uint(2), rows[1].ReferenceOne.ID)
	assert.Equal(t, uint(3), rows[2].ReferenceOne.ID)
	assert.Equal(t, uint(4), rows[3].ReferenceOne.ID)
	assert.Equal(t, uint(5), rows[4].ReferenceOne.ID)
	assert.Equal(t, "", rows[0].ReferenceOne.Name)
	assert.False(t, rows[0].ReferenceOne.Loaded())

	DBLogger := &TestDatabaseLogger{}
	pool := engine.GetMysql()
	pool.RegisterLogger(DBLogger)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)

	assert.Equal(t, uint(6), rows[0].ID)
	assert.Equal(t, uint(7), rows[1].ID)
	assert.Equal(t, uint(8), rows[2].ID)
	assert.Equal(t, uint(9), rows[3].ID)
	assert.Equal(t, uint(10), rows[4].ID)
	assert.Len(t, DBLogger.Queries, 1)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)
	assert.Equal(t, uint(6), rows[0].ID)
	assert.Equal(t, uint(7), rows[1].ID)
	assert.Equal(t, uint(8), rows[2].ID)
	assert.Equal(t, uint(9), rows[3].ID)
	assert.Equal(t, uint(10), rows[4].ID)
	assert.Len(t, DBLogger.Queries, 1)

	pager = &orm.Pager{CurrentPage: 2, PageSize: 4}
	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 1)
	assert.Equal(t, uint(10), rows[0].ID)
	assert.Len(t, DBLogger.Queries, 1)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 5}
	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)
	assert.Equal(t, uint(1), rows[0].ID)
	assert.Len(t, DBLogger.Queries, 1)

	rows[0].Age = 18
	err = rows[0].Flush()
	assert.Nil(t, err)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 10}
	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 6, totalRows)
	assert.Len(t, rows, 6)
	assert.Equal(t, uint(1), rows[0].ID)
	assert.Equal(t, uint(6), rows[1].ID)
	assert.Len(t, DBLogger.Queries, 3)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 4, totalRows)
	assert.Len(t, rows, 4)
	assert.Equal(t, uint(2), rows[0].ID)
	assert.Len(t, DBLogger.Queries, 4)

	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 10, totalRows)
	assert.Len(t, rows, 10)
	assert.Len(t, DBLogger.Queries, 5)

	rows[1].MarkToDelete()
	err = rows[1].Flush()
	assert.Nil(t, err)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 3, totalRows)
	assert.Len(t, rows, 3)
	assert.Equal(t, uint(3), rows[0].ID)
	assert.Len(t, DBLogger.Queries, 7)

	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 9, totalRows)
	assert.Len(t, rows, 9)
	assert.Len(t, DBLogger.Queries, 8)

	entity = &TestEntityIndexTestLocalRedis{Name: "Name 11", Age: uint16(18)}
	engine.RegisterEntity(entity)
	err = entity.Flush()
	assert.Nil(t, err)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 7, totalRows)
	assert.Len(t, rows, 7)
	assert.Equal(t, uint(11), rows[6].ID)
	assert.Len(t, DBLogger.Queries, 10)

	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 10, totalRows)
	assert.Len(t, rows, 10)
	assert.Len(t, DBLogger.Queries, 11)

	err = engine.ClearByIDs(entity, 1, 3)
	assert.Nil(t, err)
	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 10, totalRows)
	assert.Len(t, rows, 10)
	assert.Len(t, DBLogger.Queries, 12)

	var row TestEntityIndexTestLocalRedis
	has, err := engine.CachedSearchOne(&row, "IndexName", "Name 6")
	assert.Nil(t, err)
	assert.True(t, has)
	assert.Equal(t, uint(6), row.ID)

	has, err = engine.CachedSearchOne(&row, "IndexName", "Name 99")
	assert.Nil(t, err)
	assert.False(t, has)

	totalRows, err = engine.CachedSearchWithReferences(&rows, "IndexAll", pager, []interface{}{}, []string{"*"})
	assert.Nil(t, err)
	assert.Equal(t, 10, totalRows)
	assert.Len(t, rows, 10)
	assert.Len(t, DBLogger.Queries, 15)
	assert.Equal(t, "Name 1", rows[0].ReferenceOne.Name)
	assert.Equal(t, "Name 3", rows[1].ReferenceOne.Name)
	assert.True(t, rows[0].ReferenceOne.Loaded())
	assert.True(t, rows[1].ReferenceOne.Loaded())
}
