package gorm_reorder_test

import (
	"database/sql"
	"testing"
	"time"

	gormReorder "github.com/gopkg-dev/gorm-reorder"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string
	Email        *string
	Age          uint8
	Birthday     *time.Time
	MemberNumber sql.NullString
	ActivatedAt  sql.NullTime
}

func TestReorder(t *testing.T) {
	cfg := gormReorder.Config{
		AutoAdd:       true,
		TablePrefix:   "t_amz_",
		SingularTable: true,
	}
	reorder := gormReorder.NewReorder(cfg).AddModel([]interface{}{User{}}).Parser()
	require.NotNil(t, reorder)
	require.NotEmpty(t, reorder.GetSchemas())
	for _, schema := range reorder.GetSchemas() {
		t.Logf("%#v", schema)
	}

	marshalSchema, err := gormReorder.MarshalSchema(reorder.GetSchemas())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(marshalSchema))

}
