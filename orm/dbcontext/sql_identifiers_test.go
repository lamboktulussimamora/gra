package dbcontext

import "testing"

func TestValidateSQLIdentifierPath_AllowsSimpleAndQualified(t *testing.T) {
	cases := []string{
		"id",
		"created_at",
		"u.id",
		"user_profile.created_at",
		"_tmp",
		"t1.col2",
	}
	for _, c := range cases {
		if err := validateSQLIdentifierPath(c); err != nil {
			t.Fatalf("expected %q to be valid, got error: %v", c, err)
		}
	}
}

func TestValidateSQLIdentifierPath_RejectsSQLFragments(t *testing.T) {
	cases := []string{
		"",
		" ",
		"id desc",
		"id;drop table users",
		"id, name",
		"id) OR 1=1--",
		"1abc",
		"a-b",
		"a..b",
		"a.",
		".a",
		"a.*",
		"a(b)",
	}
	for _, c := range cases {
		if err := validateSQLIdentifierPath(c); err == nil {
			t.Fatalf("expected %q to be rejected", c)
		}
	}
}

type obUser struct{ ID int }

func TestEnhancedDbSet_OrderBySafe_DoesNotMutateOnError(t *testing.T) {
	ctx := &EnhancedDbContext{}
	set := NewEnhancedDbSet[obUser](ctx)

	if set.orderClause != "" {
		t.Fatalf("expected empty orderClause")
	}

	_, err := set.OrderBySafe("id desc")
	if err == nil {
		t.Fatalf("expected error")
	}
	if set.orderClause != "" {
		t.Fatalf("expected original set not to change")
	}
}

func TestEnhancedDbSet_OrderBySafe_SetsOrderClause(t *testing.T) {
	ctx := &EnhancedDbContext{}
	set := NewEnhancedDbSet[obUser](ctx)

	ordered, err := set.OrderBySafe("id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ordered.orderClause != "id" {
		t.Fatalf("expected orderClause to be set")
	}
}

func TestEnhancedSet_OrderBySafe_DoesNotMutateOnError(t *testing.T) {
	ctx := &EnhancedDbContext{}
	es := NewEnhancedSet[obUser](ctx)
	if len(es.builder.orderClauses) != 0 {
		t.Fatalf("expected empty orderClauses")
	}

	_, err := es.OrderBySafe("id desc")
	if err == nil {
		t.Fatalf("expected error")
	}
	if len(es.builder.orderClauses) != 0 {
		t.Fatalf("expected orderClauses not to change on error")
	}
}

func TestEnhancedSet_OrderBySafe_AppendsOnSuccess(t *testing.T) {
	ctx := &EnhancedDbContext{}
	es := NewEnhancedSet[obUser](ctx)

	_, err := es.OrderBySafe("id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(es.builder.orderClauses) != 1 {
		t.Fatalf("expected one order clause")
	}
	if es.builder.orderClauses[0].Column != "id" || es.builder.orderClauses[0].Desc {
		t.Fatalf("unexpected order clause: %+v", es.builder.orderClauses[0])
	}
}
