package tests

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type TestEntitySearch struct {
	orm.ORM
	ID           uint
	Name         string
	ReferenceOne *TestEntitySearchRef
}

type TestEntitySearchRef struct {
	orm.ORM `orm:"redisCache"`
	ID      uint
	Name    string
}

func TestSearch(t *testing.T) {
	engine := PrepareTables(t, &orm.Registry{}, TestEntitySearch{}, TestEntitySearchRef{})
	var entity TestEntitySearch

	var entities = make([]interface{}, 10)
	var refs = make([]interface{}, 10)
	for i := 1; i <= 10; i++ {
		r := &TestEntitySearchRef{Name: "Name " + strconv.Itoa(i)}
		engine.RegisterEntity(r)
		refs[i-1] = r
		err := r.Flush()
		assert.Nil(t, err)
		e := &TestEntitySearch{Name: "Name " + strconv.Itoa(i)}
		engine.RegisterEntity(e)
		e.ReferenceOne = r
		entities[i-1] = e
		err = e.Flush()
		assert.Nil(t, err)
	}
	pager := &orm.Pager{CurrentPage: 1, PageSize: 100}
	where := orm.NewWhere("`ID` > ?", 1)
	where.Append("AND `ID` < ?", 8)
	var rows []*TestEntitySearch
	err := engine.Search(where, pager, &rows, "ReferenceOne")
	assert.Nil(t, err)
	assert.True(t, rows[0].ReferenceOne.Loaded())
	assert.Len(t, rows, 6)
	assert.Equal(t, uint(2), rows[0].ID)
	assert.Equal(t, uint(7), rows[5].ID)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 4}

	rows = make([]*TestEntitySearch, 0)
	totalRows, err := engine.SearchWithCount(where, pager, &rows)
	assert.Nil(t, err)
	assert.Len(t, rows, 4)
	assert.Equal(t, 6, totalRows)
	assert.Equal(t, uint(2), rows[0].ID)
	assert.Equal(t, uint(5), rows[3].ID)

	pager = &orm.Pager{CurrentPage: 2, PageSize: 4}
	rows = make([]*TestEntitySearch, 0)
	totalRows, err = engine.SearchWithCount(where, pager, &rows)
	assert.Nil(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, 6, totalRows)
	assert.Equal(t, uint(6), rows[0].ID)
	assert.Equal(t, uint(7), rows[1].ID)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 6}
	rows = make([]*TestEntitySearch, 0)
	totalRows, err = engine.SearchWithCount(where, pager, &rows)
	assert.Nil(t, err)
	assert.Len(t, rows, 6)
	assert.Equal(t, 6, totalRows)
	assert.Equal(t, uint(2), rows[0].ID)
	assert.Equal(t, uint(7), rows[5].ID)

	pager = &orm.Pager{CurrentPage: 1, PageSize: 3}
	ids, err := engine.SearchIDs(where, pager, entity)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Equal(t, uint64(2), ids[0])

	pager.IncrementPage()
	assert.Equal(t, 2, pager.CurrentPage)
	ids, err = engine.SearchIDs(where, pager, entity)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Equal(t, uint64(5), ids[0])

	pager = &orm.Pager{CurrentPage: 1, PageSize: 100}
	rows = make([]*TestEntitySearch, 0)
	where = orm.NewWhere("(`ID` IN ? OR `ID` IN ?)", []uint{5, 6}, []uint{7, 8})
	totalRows, err = engine.SearchWithCount(where, pager, &rows)
	assert.Nil(t, err)
	assert.Len(t, rows, 4)
	assert.Equal(t, 4, totalRows)

	res, total, err := engine.SearchIDsWithCount(where, pager, entity)
	assert.Nil(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []uint64{5, 6, 7, 8}, res)
}
