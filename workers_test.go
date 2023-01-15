package main

import (
	"math/rand"
	"testing"
)

func Test_processQuery(t *testing.T) {
	type args struct {
		query            string
		randomIntPadding int64
		randomIntMax     int64
		termsMapEnabled  bool
		termsMap         map[string]string
	}
	rand.Seed(12345)
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no-replacing", args{"CREATE(n)", 0, 0, false, nil}, "CREATE(n)"},
		{"no-replacing", args{"ProblemList=[29849199,27107682]", 0, 0, false, nil}, "ProblemList=[29849199,27107682]"},
		{"no-replacing", args{"ProblemList=[29849199,__rand_int__]", 0, 1, false, nil}, "ProblemList=[29849199,0]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,__rand_int__]", 0, 1, false, nil}, "ProblemList=[0,0]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,11]", 0, 1, false, nil}, "ProblemList=[0,11]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,__rand_int__]", 0, 10, false, nil}, "ProblemList=[3,4]"},
		{"no-replacing", args{"ProblemList=[__rand_int__,__rand_int__]", -1, 10, false, nil}, "ProblemList=[7,-1]"},
		{"disabled-term-map", args{"CYPHER entityUid='__Entity__' MATCH(entity:Entity{entityUid:$entityUid}) RETURN entity", -1, 10, false, nil}, "CYPHER entityUid='__Entity__' MATCH(entity:Entity{entityUid:$entityUid}) RETURN entity"},
		{"replacing-term-map", args{"CYPHER entityUid='__Entity__' MATCH(entity:Entity{entityUid:$entityUid}) RETURN entity", -1, 10, true, map[string]string{"__Entity__": "fbfa03a5-762b-4d32-be97-f19f3f3dda72"}}, "CYPHER entityUid='fbfa03a5-762b-4d32-be97-f19f3f3dda72' MATCH(entity:Entity{entityUid:$entityUid}) RETURN entity"},
		{"term-exists-but-disabled", args{"CYPHER entityUid='__Entity__' MATCH(entity:Entity{entityUid:$entityUid}) RETURN entity", -1, 10, false, map[string]string{"__Entity__": "fbfa03a5-762b-4d32-be97-f19f3f3dda72"}}, "CYPHER entityUid='__Entity__' MATCH(entity:Entity{entityUid:$entityUid}) RETURN entity"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processQuery(tt.args.query, tt.args.randomIntPadding, tt.args.randomIntMax, tt.args.termsMapEnabled, tt.args.termsMap); got != tt.want {
				t.Errorf("processQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
