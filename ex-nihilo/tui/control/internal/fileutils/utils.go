package fileutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	DBDir = "./db"
	DBExt = ".db"
)

var illegalCharset = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func init() {
	if err := os.MkdirAll(DBDir, 0o755); err != nil {
		log.Fatalln(err.Error())
	}
}

func ExistingDBs() ([]string, error) {
	entries, err := os.ReadDir(DBDir)
	if err != nil {
		return nil, err
	}

	l := len(DBExt)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		n := e.Name()
		if !e.IsDir() && filepath.Ext(n) == DBExt {
			names = append(names, n[:len(n)-l])
		} else {
			_ = os.Remove(filepath.Join(DBDir, n))
		}
	}
	return names, nil
}

func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be blank")
	}
	if s := findIllegalChar(name); s != "" {
		return fmt.Errorf("illegal character %q", s)
	}
	if _, err := os.Stat(GetDBPath(name)); err == nil {
		return fmt.Errorf("a DB named %q already exists", name)
	}
	return nil
}

func GetDBPath(name string) string {
	return filepath.Join(DBDir, name+DBExt)
}

func findIllegalChar(s string) string {
	return illegalCharset.FindString(s)
}
