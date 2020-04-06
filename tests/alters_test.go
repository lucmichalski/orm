package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
	"testing"
	"time"
)

type AddressSchema struct {
	Street   string
	Building uint16
}

type fieldsColors struct {
	Red    string
	Green  string
	Blue   string
	Yellow string
	Purple string
}

var Color = &fieldsColors{
	Red:    "Red",
	Green:  "Green",
	Blue:   "Blue",
	Yellow: "Yellow",
	Purple: "Purple",
}

type TestEntitySchema struct {
	Orm                  *orm.ORM `orm:"mysql=schema"`
	Id                   uint
	Name                 string `orm:"length=100;index=FirstIndex"`
	NameNotNull          string `orm:"length=100;index=FirstIndex;required"`
	BigName              string `orm:"length=max"`
	Uint8                uint8  `orm:"unique=SecondIndex:2,ThirdIndex"`
	Uint24               uint32 `orm:"mediumint=true"`
	Uint32               uint32
	Uint64               uint64 `orm:"unique=SecondIndex"`
	Int8                 int8
	Int16                int16
	Int32                int32
	Int64                int64
	Rune                 rune
	Int                  int
	Bool                 bool
	Float32              float32
	Float64              float64
	Float32Decimal       float32  `orm:"decimal=8,2"`
	Float64DecimalSigned float64  `orm:"decimal=8,2;unsigned=false"`
	Enum                 string   `orm:"enum=tests.Color"`
	EnumNotNull          string   `orm:"enum=tests.Color;required"`
	Set                  []string `orm:"set=tests.Color"`
	Year                 uint16   `orm:"year=true"`
	YearNotNull          uint16   `orm:"year=true;required"`
	Date                 time.Time
	DateNotNull          time.Time `orm:"required"`
	DateTime             time.Time `orm:"time=true"`
	Address              AddressSchema
	Json                 interface{}
	ReferenceOne         *orm.ReferenceOne `orm:"ref=tests.TestEntitySchemaRef"`
	ReferenceOneCascade  *orm.ReferenceOne `orm:"ref=tests.TestEntitySchemaRef;cascade"`
	IgnoreField          []time.Time       `orm:"ignore"`
	Blob                 []byte
}

type TestEntitySchema2Entity struct {
	Orm            *orm.ORM `orm:"mysql=schema"`
	ID             uint64
	UserID         *orm.ReferenceOne `orm:"ref=tests.TestEntitySchema2EntityRef1;unique=myunique:1"`
	OrganizationID *orm.ReferenceOne `orm:"ref=tests.TestEntitySchema2EntityRef2;unique=myunique:2;index=Organization"`
	RoleID         *orm.ReferenceOne `orm:"ref=tests.TestEntitySchema2EntityRef3;unique=myunique:3"`
	FakeDelete     bool
}

type TestEntitySchema2EntityRef1 struct {
	Orm *orm.ORM `orm:"mysql=schema"`
	ID  uint64
}

type TestEntitySchema2EntityRef2 struct {
	Orm *orm.ORM `orm:"mysql=schema"`
	ID  uint64
}

type TestEntitySchema2EntityRef3 struct {
	Orm *orm.ORM `orm:"mysql=schema"`
	ID  uint64
}

type TestEntitySchemaRef struct {
	Orm  *orm.ORM `orm:"mysql=schema"`
	Id   uint
	Name string
}

func TestGetAlters(t *testing.T) {

	config := &orm.Config{}

	err := config.RegisterMySqlPool("root:root@tcp(localhost:3308)/test_schema", "schema")
	assert.Nil(t, err)
	engine := orm.NewEngine(config)

	var entity TestEntitySchema
	var entityRef TestEntitySchemaRef
	var entity2 TestEntitySchema2Entity

	var ref1 TestEntitySchema2EntityRef1
	var ref2 TestEntitySchema2EntityRef2
	var ref3 TestEntitySchema2EntityRef3

	config.RegisterEntity(entity, entityRef, entity2, ref1, ref2, ref3)
	config.RegisterEnum("tests.Color", Color)
	tableSchema, err := config.GetTableSchema(entity)
	assert.Nil(t, err)
	err = tableSchema.DropTable(engine)
	assert.Nil(t, err)
	tableSchemaRef, err := config.GetTableSchema(entityRef)
	assert.Nil(t, err)
	err = tableSchemaRef.DropTable(engine)
	assert.Nil(t, err)
	err = tableSchemaRef.DropTable(engine)
	assert.Nil(t, err)
	tableSchemaRef, err = config.GetTableSchema(entity2)
	assert.Nil(t, err)
	err = tableSchemaRef.DropTable(engine)
	assert.Nil(t, err)
	_, err = engine.GetAlters()
	assert.Nil(t, err)
}
