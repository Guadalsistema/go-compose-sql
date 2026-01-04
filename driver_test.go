package sqlcompose

import "testing"

func TestDriverByNamePostgres(t *testing.T) {
	d, _ := DriverByName("postgres")
	if _, ok := d.(PostgresDriver); !ok {
		t.Fatalf("expected PostgresDriver, got %T", d)
	}
}
