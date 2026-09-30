package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	db    *pgxpool.Pool
	cache *redis.Client
}

func main() {
	ctx := context.Background()
	dbURL := getenv("DATABASE_URL", "postgres://movank:movank_dev@localhost:5432/movank?sslmode=disable")
	cacheURL := getenv("REDIS_URL", "redis://localhost:6379")

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil { log.Fatal(err) }
	defer db.Close()

	cacheOpts, err := redis.ParseURL(cacheURL)
	if err != nil { log.Fatal(err) }
	cache := redis.NewClient(cacheOpts)
	defer cache.Close()

	app := &App{db: db, cache: cache}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.health)
	mux.HandleFunc("/v1/products", app.products)
	mux.HandleFunc("/v1/sales", app.sales)
	mux.HandleFunc("/v1/dashboard/today", app.dashboard)

	s := &http.Server{Addr: ":" + getenv("PORT", "8080"), Handler: logging(mux)}
	log.Printf("MOVANK API listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" { return v }
	return fallback
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status":"ok","time":time.Now().UTC()})
}

func (a *App) products(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, err := a.db.Query(r.Context(), `SELECT id, name, price_cents FROM products WHERE tenant_id=$1 ORDER BY name`, "demo")
		if err != nil { http.Error(w, err.Error(), 500); return }
		defer rows.Close()
		type P struct { ID uuid.UUID `json:"id"`; Name string `json:"name"`; PriceCents int64 `json:"price_cents"` }
		out := []P{}
		for rows.Next() {
			var p P
			if err := rows.Scan(&p.ID,&p.Name,&p.PriceCents); err != nil { http.Error(w, err.Error(),500); return }
			out = append(out,p)
		}
		writeJSON(w,200,out); return
	}
	if r.Method != http.MethodPost { http.Error(w,"method not allowed",405); return }
	var in struct { Name string `json:"name"`; PriceCents int64 `json:"price_cents"` }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w,"invalid json",400); return }
	var id uuid.UUID
	err := a.db.QueryRow(r.Context(), `INSERT INTO products(tenant_id,name,price_cents) VALUES($1,$2,$3) RETURNING id`, "demo", in.Name,in.PriceCents).Scan(&id)
	if err != nil { http.Error(w,err.Error(),500); return }
	writeJSON(w,201,map[string]any{"id":id,"name":in.Name,"price_cents":in.PriceCents})
}

func (a *App) sales(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w,"method not allowed",405); return }
	key := r.Header.Get("Idempotency-Key")
	if key == "" { http.Error(w,"Idempotency-Key required",400); return }

	var in struct { ProductID string `json:"product_id"`; Quantity int `json:"quantity"` }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w,"invalid json",400); return }
	qty := in.Quantity; if qty < 1 { qty = 1 }

	tx, err := a.db.Begin(r.Context()); if err != nil { http.Error(w,err.Error(),500); return }
	defer tx.Rollback()

	var existing uuid.UUID
	err = tx.QueryRow(r.Context(), `SELECT resource_id FROM idempotency_keys WHERE tenant_id=$1 AND key=$2`, "demo", key).Scan(&existing)
	if err == nil {
		tx.Commit(r.Context())
		writeJSON(w,200,map[string]any{"id":existing,"idempotent":true}); return
	}

	var product uuid.UUID
	if err := uuid.Parse(in.ProductID); err != nil { http.Error(w,"invalid product_id",400); return }
	var cents int64
	if err := tx.QueryRow(r.Context(), `SELECT id, price_cents FROM products WHERE tenant_id=$1 AND id=$2`, "demo", product).Scan(&product,&cents); err != nil {
		http.Error(w,"product not found",404); return
	}

	var saleID uuid.UUID
	if err := tx.QueryRow(r.Context(), `INSERT INTO sales(tenant_id,total_cents,status) VALUES($1,$2,'PENDING_SYNC') RETURNING id`, "demo", cents*int64(qty)).Scan(&saleID); err != nil { http.Error(w,err.Error(),500); return }
	if _,err := tx.Exec(r.Context(), `INSERT INTO idempotency_keys(tenant_id,key,resource_id) VALUES($1,$2,$3)`, "demo",key,saleID); err != nil { http.Error(w,err.Error(),500); return }
	if _,err := tx.Exec(r.Context(), `INSERT INTO outbox(tenant_id,event_type,aggregate_id,payload) VALUES($1,'sale.created',$2,$3)`, "demo",saleID,`{"status":"PENDING_SYNC"}`); err != nil { http.Error(w,err.Error(),500); return }

	if err := tx.Commit(r.Context()); err != nil { http.Error(w,err.Error(),500); return }
	writeJSON(w,201,map[string]any{"id":saleID,"status":"PENDING_SYNC","idempotent":false})
}

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const key = "dashboard:demo:today"
	if cached, err := a.cache.Get(ctx,key).Result(); err == nil {
		w.Header().Set("Content-Type","application/json"); w.Write([]byte(cached)); return
	}
	var count int64
	var total int64
	if err := a.db.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(total_cents),0) FROM sales WHERE tenant_id=$1 AND status='APPROVED'`, "demo").Scan(&count,&total); err != nil { http.Error(w,err.Error(),500); return }
	body := map[string]any{"sales_count":count,"total_cents":total}
	raw,_ := json.Marshal(body)
	_ = a.cache.Set(ctx,key,raw,30*time.Second).Err()
	writeJSON(w,200,body)
}

func writeJSON(w http.ResponseWriter, code int, v any) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(code); _=json.NewEncoder(w).Encode(v) }
func logging(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ log.Printf("%s %s",r.Method,r.URL.Path); next.ServeHTTP(w,r) }) }
