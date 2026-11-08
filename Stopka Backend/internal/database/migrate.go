package database

import (
 "context"
 "fmt"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "time"
 "github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(pool *pgxpool.Pool) error {
 ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
 defer cancel()
 if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil { return fmt.Errorf("таблица миграций: %w", err) }
 entries, err := os.ReadDir("migrations")
 if err != nil { return fmt.Errorf("чтение папки migrations: %w", err) }
 var files []string
 for _, e := range entries { if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") { files = append(files, e.Name()) } }
 sort.Strings(files)
 for _, name := range files {
   tx, err := pool.Begin(ctx)
   if err != nil { return fmt.Errorf("транзакция %s: %w", name, err) }
   var exists bool
   err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename=$1)`, name).Scan(&exists)
   if err != nil { tx.Rollback(ctx); return err }
   if exists { tx.Rollback(ctx); continue }
   sql, err := os.ReadFile(filepath.Join("migrations", name))
   if err != nil { tx.Rollback(ctx); return err }
   if _, err = tx.Exec(ctx, string(sql)); err != nil { tx.Rollback(ctx); return fmt.Errorf("миграция %s: %w", name, err) }
   if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(filename) VALUES ($1)`, name); err != nil { tx.Rollback(ctx); return err }
   if err = tx.Commit(ctx); err != nil { return err }
   fmt.Println("Применена миграция:", name)
 }
 return nil
}
