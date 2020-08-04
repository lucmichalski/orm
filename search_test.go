package orm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type searchEntity struct {
	ORM          `orm:"localCache;redisCache"`
	ID           uint
	Name         string
	ReferenceOne *searchEntityReference
	FakeDelete   bool
}

type searchEntityReference struct {
	ORM
	ID   uint
	Name string
}

func TestSearch(t *testing.T) {
	var entity *searchEntity
	var reference *searchEntityReference
	engine := PrepareTables(t, &Registry{}, entity, reference)

	for i := 1; i <= 10; i++ {
		engine.Track(&searchEntity{Name: fmt.Sprintf("name %d", i), ReferenceOne: &searchEntityReference{Name: fmt.Sprintf("name %d", i)}})
	}
	engine.Flush()

	var rows []*searchEntity
	missing := engine.LoadByIDs([]uint64{1, 2, 20}, &rows)
	assert.Len(t, missing, 1)
	assert.Len(t, rows, 2)
	assert.Equal(t, uint64(20), missing[0])
	assert.Equal(t, uint(1), rows[0].ID)
	assert.Equal(t, uint(2), rows[1].ID)

	entity = &searchEntity{}
	found := engine.SearchOne(NewWhere("ID = ?", 1), entity, "ReferenceOne")
	assert.True(t, found)
	assert.Equal(t, uint(1), entity.ID)
	assert.Equal(t, "name 1", entity.Name)
	assert.Equal(t, "name 1", entity.ReferenceOne.Name)
	assert.True(t, engine.Loaded(entity.ReferenceOne))

	engine.Search(NewWhere("ID > 0"), nil, &rows, "ReferenceOne")
	assert.Len(t, rows, 10)
	assert.Equal(t, uint(1), rows[0].ID)
	assert.Equal(t, "name 1", rows[0].Name)
	assert.Equal(t, "name 1", rows[0].ReferenceOne.Name)
	assert.True(t, engine.Loaded(rows[0].ReferenceOne))

	engine = PrepareTables(t, &Registry{})
	assert.PanicsWithValue(t, EntityNotRegisteredError{Name: "orm.searchEntity"}, func() {
		engine.Search(NewWhere("ID > 0"), nil, &rows)
	})
}
