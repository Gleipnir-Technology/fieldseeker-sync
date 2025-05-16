package fssync

import (
	"testing"

	"github.com/Gleipnir-Technology/arcgis-go"
)

func TestUpsertFromQueryResult(t *testing.T) {
	qr := arcgis.QueryResult{
		Fields: []arcgis.Field{
			{Name: "OBJECTID"},
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
		ObjectIdFieldName: "OBJECTID",
		UniqueIdField: arcgis.UniqueIdField{
			Name: "OBJECTID",
		},
	}
	query := upsertFromQueryResult("foo", &qr)
	if query != `INSERT INTO foo (OBJECTID,a,b,c,geometry_x,geometry_y)
VALUES (@OBJECTID,@a,@b,@c,@geometry_x,@geometry_y)
ON CONFLICT(OBJECTID)
DO UPDATE SET
 a = EXCLUDED.a,
 b = EXCLUDED.b,
 c = EXCLUDED.c,
 geometry_x = EXCLUDED.geometry_x,
 geometry_y = EXCLUDED.geometry_y
;` {
		t.Errorf("Got wrong query: %v", query)
	}
}
