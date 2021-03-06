package orm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type lazyReceiverEntity struct {
	ORM          `orm:"localCache;redisCache"`
	ID           uint
	Name         string `orm:"unique=name"`
	Age          uint64
	EnumNullable string `orm:"enum=orm.TestEnum"`
	RefOne       *lazyReceiverReference
	IndexAll     *CachedQuery `query:""`
}

type lazyReceiverReference struct {
	ORM
	ID   uint
	Name string
}

func TestLazyReceiver(t *testing.T) {
	var entity *lazyReceiverEntity
	var ref *lazyReceiverReference

	registry := &Registry{}
	registry.RegisterEnumMap("orm.TestEnum", map[string]string{"a": "a", "b": "b", "c": "c"}, "a")
	engine := PrepareTables(t, registry, entity, ref)

	receiver := NewLazyReceiver(engine)
	receiver.DisableLoop()
	receiver.SetMaxLoopDuration(time.Millisecond)
	receiver.Purge()

	e := &lazyReceiverEntity{Name: "John", Age: 18}
	engine.Track(e)
	engine.FlushLazy()

	e = &lazyReceiverEntity{}
	loaded := engine.LoadByID(1, e)
	assert.False(t, loaded)

	validHeartBeat := false
	receiver.SetHeartBeat(func() {
		validHeartBeat = true
	})
	receiver.Digest()
	assert.True(t, validHeartBeat)

	loaded = engine.LoadByID(1, e)
	assert.True(t, loaded)
	assert.Equal(t, "John", e.Name)
	assert.Equal(t, uint64(18), e.Age)

	e.Name = "Tom"
	engine.Track(e)
	engine.FlushLazy()

	e = &lazyReceiverEntity{}
	loaded = engine.LoadByID(1, e)
	assert.True(t, loaded)
	assert.Equal(t, "Tom", e.Name)

	engine.GetLocalCache().Clear()
	e = &lazyReceiverEntity{}
	loaded = engine.LoadByID(1, e)
	assert.True(t, loaded)
	assert.Equal(t, "John", e.Name)

	validHeartBeat = false
	receiver.SetHeartBeat(func() {
		validHeartBeat = true
	})
	receiver.Digest()
	assert.True(t, validHeartBeat)

	e = &lazyReceiverEntity{}
	loaded = engine.LoadByID(1, e)
	assert.True(t, loaded)
	assert.Equal(t, "Tom", e.Name)

	e = &lazyReceiverEntity{Name: "Tom"}
	engine.Track(e)
	engine.FlushLazy()

	validHeartBeat = false
	receiver.SetHeartBeat(func() {
		validHeartBeat = true
	})
	receiver.Digest()
	assert.True(t, validHeartBeat)

	e = &lazyReceiverEntity{Name: "Adam", EnumNullable: "wrong"}
	engine.Track(e)
	engine.FlushLazy()

	assert.NotPanics(t, func() {
		receiver.Digest()
	})
	e = &lazyReceiverEntity{Name: "Tom"}
	engine.SetOnDuplicateKeyUpdate(NewWhere("Age = ?", 38), e)
	engine.Track(e)
	assert.PanicsWithError(t, "lazy flush on duplicate key not supported", func() {
		engine.FlushLazy()
	})
	engine.ClearTrackedEntities()

	e = &lazyReceiverEntity{Name: "Adam", RefOne: &lazyReceiverReference{Name: "Test"}}
	engine.Track(e)
	assert.PanicsWithError(t, "lazy flush for unsaved references not supported", func() {
		engine.FlushLazy()
	})
	engine.ClearTrackedEntities()

	e = &lazyReceiverEntity{}
	engine.LoadByID(1, e)
	engine.MarkToDelete(e)
	engine.FlushLazy()
	receiver.Digest()
	loaded = engine.LoadByID(1, e)
	assert.False(t, loaded)
}
