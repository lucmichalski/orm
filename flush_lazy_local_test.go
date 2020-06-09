package orm

import (
	"strconv"
	"testing"

	log2 "github.com/apex/log"

	"github.com/apex/log/handlers/memory"

	"github.com/stretchr/testify/assert"
)

type testEntityFlushLazyLocal struct {
	ORM  `orm:"localCache"`
	ID   uint
	Name string
}

func TestFlushLazyLocal(t *testing.T) {
	var entity testEntityFlushLazyLocal
	engine := PrepareTables(t, &Registry{}, entity)
	DBLogger := memory.New()
	engine.AddQueryLogger(DBLogger, log2.InfoLevel, QueryLoggerSourceDB)

	var entities = make([]interface{}, 10)
	for i := 1; i <= 10; i++ {
		e := &testEntityFlushLazyLocal{Name: "Name " + strconv.Itoa(i)}
		entities[i-1] = e
		engine.Track(e)
	}
	engine.FlushLazy()
	assert.Len(t, DBLogger.Entries, 0)

	receiver := NewLazyReceiver(engine)
	receiver.DisableLoop()
	receiver.Digest()
	assert.Len(t, DBLogger.Entries, 1)
	found := engine.LoadByID(1, &entity)
	assert.True(t, found)
	assert.Equal(t, "Name 1", entity.Name)

	DBLogger.Entries = make([]*log2.Entry, 0)
	engine.Track(&entity)
	entity.Name = "Name 1.1"
	engine.FlushLazy()
	assert.Len(t, DBLogger.Entries, 0)

	receiver.Digest()
	assert.Len(t, DBLogger.Entries, 1)
	found = engine.LoadByID(1, &entity)
	assert.True(t, found)
	assert.Equal(t, "Name 1.1", entity.Name)

	DBLogger.Entries = make([]*log2.Entry, 0)
	engine.MarkToDelete(&entity)
	engine.FlushLazy()
	assert.Len(t, DBLogger.Entries, 0)

	receiver.Digest()
	assert.Len(t, DBLogger.Entries, 1)
	found = engine.LoadByID(1, &entity)
	assert.False(t, found)
}
