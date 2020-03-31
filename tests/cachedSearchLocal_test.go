package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
	"strconv"
	"testing"
)

type TestEntityIndexTestLocal struct {
	Orm          *orm.ORM `orm:"localCache"`
	Id           uint
	Name         string `orm:"length=100;index=FirstIndex"`
	Age          uint16
	Ignore       uint16            `orm:"ignore"`
	IndexAge     *orm.CachedQuery  `query:":Age = ? ORDER BY :Id"`
	IndexAll     *orm.CachedQuery  `query:""`
	IndexName    *orm.CachedQuery  `queryOne:":Name = ?"`
	ReferenceOne *orm.ReferenceOne `orm:"ref=tests.TestEntityIndexTestLocalRef"`
}

type TestEntityIndexTestLocalRef struct {
	Orm  *orm.ORM
	Id   uint
	Name string
}

func TestCachedSearchLocal(t *testing.T) {
	var entity TestEntityIndexTestLocal
	var entityRef TestEntityIndexTestLocalRef
	engine := PrepareTables(t, &orm.Config{}, entityRef, entity)

	for i := 1; i <= 5; i++ {
		e := &TestEntityIndexTestLocalRef{Name: "Name " + strconv.Itoa(i)}
		err := engine.Flush(e)
		assert.Nil(t, err)
	}

	var entities = make([]interface{}, 10)
	for i := 1; i <= 5; i++ {
		e := TestEntityIndexTestLocal{Name: "Name " + strconv.Itoa(i), Age: uint16(10)}
		err := engine.Init(&e)
		assert.Nil(t, err)
		e.ReferenceOne.Id = uint64(i)
		entities[i-1] = &e
	}
	for i := 6; i <= 10; i++ {
		e := TestEntityIndexTestLocal{Name: "Name " + strconv.Itoa(i), Age: uint16(18)}
		entities[i-1] = &e
	}

	err := engine.Flush(entities...)
	assert.Nil(t, err)

	pager := &orm.Pager{CurrentPage: 1, PageSize: 100}
	var rows []*TestEntityIndexTestLocal
	totalRows, err := engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)
	assert.Equal(t, uint64(1), rows[0].ReferenceOne.Id)
	assert.Equal(t, uint64(2), rows[1].ReferenceOne.Id)
	assert.Equal(t, uint64(3), rows[2].ReferenceOne.Id)
	assert.Equal(t, uint64(4), rows[3].ReferenceOne.Id)
	assert.Equal(t, uint64(5), rows[4].ReferenceOne.Id)

	DBLogger := &TestDatabaseLogger{}
	engine.GetMysql().RegisterLogger(DBLogger.Logger())

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)

	assert.Equal(t, uint(6), rows[0].Id)
	assert.Equal(t, uint(7), rows[1].Id)
	assert.Equal(t, uint(8), rows[2].Id)
	assert.Equal(t, uint(9), rows[3].Id)
	assert.Equal(t, uint(10), rows[4].Id)
	assert.Len(t, DBLogger.Queries, 1)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)
	assert.Equal(t, uint(6), rows[0].Id)
	assert.Equal(t, uint(7), rows[1].Id)
	assert.Equal(t, uint(8), rows[2].Id)
	assert.Equal(t, uint(9), rows[3].Id)
	assert.Equal(t, uint(10), rows[4].Id)
	assert.Len(t, DBLogger.Queries, 1)

	pager = &orm.Pager{CurrentPage: 2, PageSize: 4}
	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 1)
	assert.Equal(t, uint(10), rows[0].Id)
	assert.Len(t, DBLogger.Queries, 1)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 5}
	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 5, totalRows)
	assert.Len(t, rows, 5)
	assert.Equal(t, uint(1), rows[0].Id)
	assert.Len(t, DBLogger.Queries, 1)

	rows[0].Age = 18
	err = engine.Flush(rows[0])
	assert.Nil(t, err)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 10}
	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 6, totalRows)
	assert.Len(t, rows, 6)
	assert.Equal(t, uint(1), rows[0].Id)
	assert.Equal(t, uint(6), rows[1].Id)
	assert.Len(t, DBLogger.Queries, 3)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 4, totalRows)
	assert.Len(t, rows, 4)
	assert.Equal(t, uint(2), rows[0].Id)
	assert.Len(t, DBLogger.Queries, 4)

	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 10, totalRows)
	assert.Len(t, rows, 10)
	assert.Len(t, DBLogger.Queries, 5)

	rows[1].Orm.MarkToDelete()
	err = engine.Flush(rows[1])
	assert.Nil(t, err)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 10)
	assert.Nil(t, err)
	assert.Equal(t, 3, totalRows)
	assert.Len(t, rows, 3)
	assert.Equal(t, uint(3), rows[0].Id)
	assert.Len(t, DBLogger.Queries, 7)

	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 9, totalRows)
	assert.Len(t, rows, 9)
	assert.Len(t, DBLogger.Queries, 8)

	entity = TestEntityIndexTestLocal{Name: "Name 11", Age: uint16(18)}
	err = engine.Flush(&entity)
	assert.Nil(t, err)

	totalRows, err = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	assert.Nil(t, err)
	assert.Equal(t, 7, totalRows)
	assert.Len(t, rows, 7)
	assert.Equal(t, uint(11), rows[6].Id)
	assert.Len(t, DBLogger.Queries, 10)

	totalRows, err = engine.CachedSearch(&rows, "IndexAll", pager)
	assert.Nil(t, err)
	assert.Equal(t, 10, totalRows)
	assert.Len(t, rows, 10)
	assert.Len(t, DBLogger.Queries, 11)

	var row TestEntityIndexTestLocal
	has, err := engine.CachedSearchOne(&row, "IndexName", "Name 6")
	assert.Nil(t, err)
	assert.True(t, has)
	assert.Equal(t, uint(6), row.Id)

	has, err = engine.CachedSearchOne(&row, "IndexName", "Name 99")
	assert.Nil(t, err)
	assert.False(t, has)

}

func BenchmarkCachedSearchLocal(b *testing.B) {
	var entity TestEntityIndexTestLocal
	var entityRef TestEntityIndexTestLocalRef
	engine := PrepareTables(&testing.T{}, &orm.Config{}, entity, entityRef)

	var entities = make([]interface{}, 10)
	for i := 1; i <= 10; i++ {
		e := TestEntityIndexTestLocal{Name: "Name " + strconv.Itoa(i), Age: uint16(18)}
		entities[i-1] = &e
	}
	_ = engine.Flush(entities...)
	pager := &orm.Pager{CurrentPage: 1, PageSize: 100}
	var rows []*TestEntityIndexTestLocal
	for n := 0; n < b.N; n++ {
		_, _ = engine.CachedSearch(&rows, "IndexAge", pager, 18)
	}
}
