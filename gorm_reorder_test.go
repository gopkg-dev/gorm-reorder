package gorm_reorder

import (
	"database/sql"
	"testing"
	"time"

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
	cfg := Config{
		AutoAdd:       true,
		TablePrefix:   "t_amz_",
		SingularTable: true,
	}
	reorder := NewReorder(cfg).AddModel([]interface{}{User{}}).Parser()
	require.NotNil(t, reorder)
	require.NotEmpty(t, reorder.GetSchemas())
	for _, schema := range reorder.GetSchemas() {
		t.Logf("%#v", schema)
	}
}
