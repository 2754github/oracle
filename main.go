package main

import (
	"bufio"
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB_FILE_PATH   = "./oracle.db"
	INDEX_DIR_PATH = "./memo"
	INDEX_FILE_EXT = []string{".md"}

	SQL_PRAGMA = `
		PRAGMA foreign_keys=true;
	`
	SQL_CREATE_TABLES = `
		CREATE TABLE note (
			id          INTEGER PRIMARY KEY,
			title       TEXT    NOT NULL,
			path        TEXT    NOT NULL,
			modified_at INTEGER NOT NULL
		);

		CREATE TABLE tag (
			id      INTEGER PRIMARY KEY,
			name    TEXT    NOT NULL UNIQUE
		);

		CREATE TABLE note_tag (
			note_id INTEGER REFERENCES note(id),
			tag_id  INTEGER REFERENCES tag(id)
		);
	`
)

func main() {
	_ = os.Remove(DB_FILE_PATH)

	db, err := sql.Open("sqlite3", DB_FILE_PATH)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(SQL_PRAGMA + SQL_CREATE_TABLES)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(INDEX_DIR_PATH, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !slices.Contains(INDEX_FILE_EXT, filepath.Ext(path)) {
			return nil
		}

		n, err := newNote(path, info.ModTime())
		if err != nil {
			return err
		}

		noteRow, err := db.Exec("INSERT INTO note(title, path, modified_at) VALUES (?, ?, ?)", n.title, n.path, n.modifiedAt)
		if err != nil {
			return err
		}

		noteID, err := noteRow.LastInsertId()
		if err != nil {
			return err
		}

		for _, tagName := range n.tags {
			var tagID int64
			err := db.QueryRow("SELECT id FROM tag WHERE name = ?", tagName).Scan(&tagID)
			if err != nil {
				if err != sql.ErrNoRows {
					return err
				}

				tagRow, err := db.Exec("INSERT INTO tag(name) VALUES (?)", tagName)
				if err != nil {
					return err
				}

				tagID, err = tagRow.LastInsertId()
				if err != nil {
					return err
				}
			}

			_, err = db.Exec("INSERT INTO note_tag(note_id, tag_id) VALUES (?, ?)", noteID, tagID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

type note struct {
	title      string
	path       string
	modifiedAt int64
	tags       []string
}

func newNote(path string, modTime time.Time) (*note, error) {
	s := strings.Split(path[:len(path)-len(filepath.Ext(path))], "/")
	title := s[len(s)-1]

	fm, err := newFrontMatter(path)
	if err != nil {
		return nil, err
	}

	return &note{
		title:      title,
		path:       path,
		modifiedAt: modTime.Unix(),
		tags:       fm.Tags,
	}, nil
}

type FrontMatter struct {
	Tags []string `yaml:"tags"`
}

func newFrontMatter(path string) (*FrontMatter, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if l == "" { // NOTE: Front matter and main content should be separated by one or more blank lines.
			break
		}
		s += l + "\n"
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	fm := new(FrontMatter)
	_, err = frontmatter.Parse(strings.NewReader(s), fm)
	if err != nil {
		return nil, err
	}

	return fm, nil
}
