package main

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
	if query != `INSERT INTO foo (OBJECTID,a,b,c)
VALUES (@OBJECTID,@a,@b,@c)
ON CONFLICT(OBJECTID)
DO UPDATE SET
 a = EXCLUDED.a,
 b = EXCLUDED.b,
 c = EXCLUDED.c
;` {
		t.Errorf("Got wrong query: %v", query)
	}
}
