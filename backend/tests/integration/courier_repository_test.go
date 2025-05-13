package integration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"backend/internal/repository"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func TestCourierRepository_FindNearestAvailable(t *testing.T) {
	// 1) Поднимаем PostGIS
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %v", err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgis/postgis",
		Tag:        "15-3.3",
		Env: []string{
			"POSTGRES_USER=user",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=tracking_service_test",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("could not start resource: %v", err)
	}
	defer pool.Purge(resource)

	// 2) Ждём доступности БД
	var db *sql.DB
	pool.MaxWait = 2 * time.Minute
	port := resource.GetPort("5432/tcp")
	dsn := fmt.Sprintf(
		"postgres://user:password@localhost:%s/tracking_service_test?sslmode=disable",
		port,
	)
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}

	// 3) Миграция
	migrationPath := filepath.Join("..", "..", "migrations", "000001_create_initial_schema_with_postgis.up.sql")
	sqlBytes, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("could not read migration file %q: %v", migrationPath, err)
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		t.Fatalf("could not apply migration: %v", err)
	}

	// 4) Сидим двух курьеров
	now := time.Now().UTC()
	_, err = db.Exec(`
		INSERT INTO users (id,email,password_hash,role,created_at,updated_at)
		VALUES ($1,'a@a.com','','COURIER',$2,$2),($3,'b@b.com','','COURIER',$2,$2)
	`, uuid.MustParse("11111111-1111-1111-1111-111111111111"), now,
		uuid.MustParse("22222222-2222-2222-2222-222222222222"))
	if err != nil {
		t.Fatalf("could not seed users: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO couriers (user_id,name,status,location,rating)
		VALUES
		  ('11111111-1111-1111-1111-111111111111','Alice','AVAILABLE',
		    ST_SetSRID(ST_MakePoint(4.90,52.37),4326),4.9),
		  ('22222222-2222-2222-2222-222222222222','Bob','AVAILABLE',
		    ST_SetSRID(ST_MakePoint(4.95,52.35),4326),4.7)
	`)
	if err != nil {
		t.Fatalf("could not seed couriers: %v", err)
	}

	// 5) Проверяем FindNearestAvailable
	repo := repository.NewCourierRepository(db, zap.NewNop())
	couriers, err := repo.FindNearestAvailable(52.37, 4.90, 5000)
	if err != nil {
		t.Fatalf("FindNearestAvailable() error: %v", err)
	}
	if len(couriers) != 2 {
		t.Errorf("len = %d; want 2", len(couriers))
	}
	if couriers[0].UserID.String() != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("first courier = %s; want Alice", couriers[0].UserID)
	}
}
