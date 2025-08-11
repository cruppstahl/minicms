package core

import "time"

type FileMetadata struct {
	Title            string    `yaml:"title"`
	Author           string    `yaml:"author"`
	CssFile          string    `yaml:"css-file"`
	Tags             []string  `yaml:"tags"`
	MimeType         string    `yaml:"mime-type"`
	RedirectUrl      string    `yaml:"redirect-url"`
	IgnoreLayout     bool      `yaml:"ignore-layout"`
	DateOfLastUpdate time.Time `yaml:"date-of-last-update"`
}

type DirectoryMetadata struct {
	Title   string `yaml:"title"`
	CssFile string `yaml:"css-file"`
}
