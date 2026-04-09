package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Reading struct {
	ID         int64
	Systolic   int
	Diastolic  int
	Pulse      int
	RecordedAt time.Time
}

var DB *sql.DB

func Init() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(dir, "BPlogger")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(appDir, "bplogger.db")
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	return migrate()
}

func migrate() error {
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS readings (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		systolic    INTEGER NOT NULL,
		diastolic   INTEGER NOT NULL,
		pulse       INTEGER NOT NULL,
		recorded_at TEXT    NOT NULL,
		tags        TEXT    NOT NULL DEFAULT '',
		notes       TEXT
	)`)
	if err != nil {
		return err
	}
	// add tags column to existing databases that predate this field
	_, _ = DB.Exec(`ALTER TABLE readings ADD COLUMN tags TEXT NOT NULL DEFAULT ''`)
	return nil
}

func InsertReading(r Reading) error {
	_, err := DB.Exec(
		`INSERT INTO readings (systolic, diastolic, pulse, recorded_at) VALUES (?, ?, ?, ?)`,
		r.Systolic, r.Diastolic, r.Pulse, r.RecordedAt.Format(time.RFC3339),
	)
	return err
}

func GetReadings() ([]Reading, error) {
	rows, err := DB.Query(`SELECT id, systolic, diastolic, pulse, recorded_at FROM readings ORDER BY recorded_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []Reading
	for rows.Next() {
		var r Reading
		var ts string
		if err := rows.Scan(&r.ID, &r.Systolic, &r.Diastolic, &r.Pulse, &ts); err != nil {
			return nil, err
		}
		r.RecordedAt, _ = time.Parse(time.RFC3339, ts)
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

func DeleteReading(id int64) error {
	_, err := DB.Exec(`DELETE FROM readings WHERE id = ?`, id)
	return err
}
