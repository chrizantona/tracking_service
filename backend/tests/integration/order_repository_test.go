package integration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"backend/internal/entity"
	"backend/internal/repository"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func TestOrderRepository_CRUD(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("docker connect: %v", err)
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
		t.Fatalf("start resource: %v", err)
	}
	defer pool.Purge(resource)

	var db *sql.DB
	pool.MaxWait = 2 * time.Minute
	dsn := fmt.Sprintf("postgres://user:password@localhost:%s/tracking_service_test?sslmode=disable", resource.GetPort("5432/tcp"))
	if err := pool.Retry(func() error {
		var e error
		db, e = sql.Open("postgres", dsn)
		if e != nil {
			return e
		}
		return db.Ping()
	}); err != nil {
		t.Fatalf("db ping: %v", err)
	}

	migrationPath := filepath.Join("..", "..", "migrations", "000001_create_initial_schema_with_postgis.up.sql")
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	now := time.Now().UTC()
	clientID := uuid.New()
	_, err = db.Exec(`INSERT INTO users (id,email,password_hash,role,created_at,updated_at) VALUES ($1,'client@example.com','', 'CLIENT',$2,$2)`, clientID, now)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, _ = db.Exec(`INSERT INTO clients (user_id,name,phone) VALUES ($1,'Test Client','+10000000000')`, clientID)

	repo := repository.NewOrderRepository(db, zap.NewNop())

	order := &entity.Order{
		ClientID:        clientID,
		Status:          entity.StatusCreated,
		DeliveryAddress: "123 Test St",
		DeliveryCoords:  "10.0,20.0",
	}
	if err := repo.Create(order); err != nil {
		t.Fatalf("Create(): %v", err)
	}
	if order.ID == uuid.Nil {
		t.Fatal("ID not set after Create()")
	}

	got, err := repo.GetByID(order.ID)
	if err != nil {
		t.Fatalf("GetByID(): %v", err)
	}
	if got.ClientID != clientID {
		t.Fatalf("clientID mismatch")
	}

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll(): %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("GetAll() len=%d want 1", len(all))
	}

	got.Status = entity.StatusAssigned
	got.DeliveryAddress = "456 New St"
	got.DeliveryCoords = "30.0,40.0"
	if err := repo.Update(got); err != nil {
		t.Fatalf("Update(): %v", err)
	}
	updated, _ := repo.GetByID(order.ID)
	if updated.Status != entity.StatusAssigned {
		t.Fatalf("status not updated")
	}

	if err := repo.Delete(order.ID); err != nil {
		t.Fatalf("Delete(): %v", err)
	}
	if _, err := repo.GetByID(order.ID); err == nil {
		t.Fatalf("expected error after delete")
	}
}
