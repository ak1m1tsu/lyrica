package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ak1m1tsu/lyrica/internal/domain"
)

var (
	sqliteTrack1 = domain.Track{ID: 1, TrackName: "Song A", ArtistName: "Artist X", AlbumName: "Album 1", Duration: 180}
	sqliteTrack2 = domain.Track{ID: 2, TrackName: "Song B", ArtistName: "Artist Y", AlbumName: "Album 2", Duration: 240, Instrumental: true}
)

func newTestSQLiteStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	s, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	t.Cleanup(func() { s.Close() }) //nolint:errcheck
	return s
}

func TestNewSQLiteStore(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	s.Close() //nolint:errcheck
}

func TestSQLiteStore_GetAll_Empty(t *testing.T) {
	s := newTestSQLiteStore(t)
	got := s.GetAll()
	if got == nil {
		t.Fatal("GetAll must return non-nil slice on empty store")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 tracks, got %d", len(got))
	}
}

func TestSQLiteStore_Add_and_GetAll(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	if err := s.Add(ctx, sqliteTrack1); err != nil {
		t.Fatalf("Add track1: %v", err)
	}
	if err := s.Add(ctx, sqliteTrack2); err != nil {
		t.Fatalf("Add track2: %v", err)
	}

	got := s.GetAll()
	if len(got) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(got))
	}
	if got[0].ID != sqliteTrack1.ID || got[0].TrackName != sqliteTrack1.TrackName {
		t.Fatalf("first track mismatch: got %+v", got[0])
	}
	if got[1].ID != sqliteTrack2.ID || got[1].TrackName != sqliteTrack2.TrackName {
		t.Fatalf("second track mismatch: got %+v", got[1])
	}
}

func TestSQLiteStore_Add_Deduplication(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	if err := s.Add(ctx, sqliteTrack1); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if err := s.Add(ctx, sqliteTrack1); err != nil {
		t.Fatalf("duplicate Add must not return error, got: %v", err)
	}

	if got := s.GetAll(); len(got) != 1 {
		t.Fatalf("expected 1 track after duplicate Add, got %d", len(got))
	}
}

func TestSQLiteStore_Has(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	if s.Has(sqliteTrack1.ID) {
		t.Fatal("Has must return false before Add")
	}

	s.Add(ctx, sqliteTrack1) //nolint:errcheck

	if !s.Has(sqliteTrack1.ID) {
		t.Fatal("Has must return true after Add")
	}

	s.Remove(ctx, sqliteTrack1.ID) //nolint:errcheck

	if s.Has(sqliteTrack1.ID) {
		t.Fatal("Has must return false after Remove")
	}
}

func TestSQLiteStore_Remove(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	if err := s.Remove(ctx, 999); err != nil {
		t.Fatalf("Remove non-existent must not return error, got: %v", err)
	}

	s.Add(ctx, sqliteTrack1) //nolint:errcheck
	s.Add(ctx, sqliteTrack2) //nolint:errcheck

	if err := s.Remove(ctx, sqliteTrack1.ID); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	got := s.GetAll()
	if len(got) != 1 || got[0].ID != sqliteTrack2.ID {
		t.Fatalf("expected only track2 after removing track1, got %+v", got)
	}
}

func TestSQLiteStore_GetDir(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer s.Close() //nolint:errcheck

	if got := s.GetDir(); got != dir {
		t.Fatalf("expected dir %q, got %q", dir, got)
	}
}

func TestSQLiteStore_SetDir(t *testing.T) {
	dir := t.TempDir()
	newDir := t.TempDir()

	s, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer s.Close() //nolint:errcheck

	ctx := context.Background()
	s.Add(ctx, sqliteTrack1) //nolint:errcheck
	s.Add(ctx, sqliteTrack2) //nolint:errcheck

	if err := s.SetDir(ctx, newDir); err != nil {
		t.Fatalf("SetDir: %v", err)
	}

	if got := s.GetDir(); got != newDir {
		t.Fatalf("expected new dir %q, got %q", newDir, got)
	}
	if _, err := os.Stat(filepath.Join(newDir, dbFileName)); err != nil {
		t.Fatalf("expected DB file in new dir: %v", err)
	}

	got := s.GetAll()
	if len(got) != 2 {
		t.Fatalf("expected 2 tracks after SetDir, got %d", len(got))
	}
}

func TestSQLiteStore_ContextCancelled(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := s.Add(ctx, sqliteTrack1); err == nil {
		t.Fatal("Add with cancelled context must return error")
	}
	if err := s.Remove(ctx, 1); err == nil {
		t.Fatal("Remove with cancelled context must return error")
	}
	if err := s.SetDir(ctx, t.TempDir()); err == nil {
		t.Fatal("SetDir with cancelled context must return error")
	}
}

func TestSQLiteStore_RoundTrip_AllFields(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	want := domain.Track{
		ID:           42,
		TrackName:    "Full Track",
		ArtistName:   "Full Artist",
		AlbumName:    "Full Album",
		Duration:     300.5,
		Instrumental: true,
		PlainLyrics:  "line one\nline two",
		SyncedLyrics: "[00:01.00]line one\n[00:05.00]line two",
	}

	s.Add(ctx, want) //nolint:errcheck

	got := s.GetAll()
	if len(got) != 1 {
		t.Fatalf("expected 1 track, got %d", len(got))
	}
	tr := got[0]
	if tr.ID != want.ID || tr.TrackName != want.TrackName || tr.ArtistName != want.ArtistName ||
		tr.AlbumName != want.AlbumName || tr.Duration != want.Duration ||
		tr.Instrumental != want.Instrumental || tr.PlainLyrics != want.PlainLyrics ||
		tr.SyncedLyrics != want.SyncedLyrics {
		t.Fatalf("round-trip mismatch:\n  got  %+v\n  want %+v", tr, want)
	}
}

func TestSQLiteStore_GetAll_IsolatesCallerMutation(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()
	s.Add(ctx, sqliteTrack1) //nolint:errcheck

	first := s.GetAll()
	first[0].TrackName = "mutated"

	second := s.GetAll()
	if second[0].TrackName == "mutated" {
		t.Fatal("GetAll must return an isolated copy — internal state was mutated by caller")
	}
}

func TestSQLiteStore_Close_Idempotent(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("second Close must be a no-op, got: %v", err)
	}
}

func TestSQLiteStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	s1, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	s1.Add(ctx, sqliteTrack1) //nolint:errcheck
	s1.Add(ctx, sqliteTrack2) //nolint:errcheck
	s1.Close()                //nolint:errcheck

	s2, err := NewSQLiteStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStore reopen: %v", err)
	}
	defer s2.Close() //nolint:errcheck

	got := s2.GetAll()
	if len(got) != 2 {
		t.Fatalf("expected 2 tracks after reopen, got %d", len(got))
	}
	if !s2.Has(sqliteTrack1.ID) || !s2.Has(sqliteTrack2.ID) {
		t.Fatal("in-memory index not rebuilt correctly after reopen")
	}
}
