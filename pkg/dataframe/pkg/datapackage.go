package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/slshen/paperscore/pkg/dataframe"
)

type License string

const (
	CopyrightAuthors = License("copyright-authors")
)

type DataPackage struct {
	ID        string
	Title     string
	Licenses  []License
	Resources []Resource
}

type Resource interface {
	GetDescription() string
	GetPath() string
	WriteContent(w io.Writer) error
}

type DataResource struct {
	Path        string
	Description string
	*dataframe.Data
}

type FileResource struct {
	Path        string
	Description string
	LocalPath   string
}

func (r *DataResource) GetPath() string {
	return r.Path
}

func (r *DataResource) GetDescription() string {
	return r.Description
}

func (r *DataResource) WriteContent(w io.Writer) error {
	return r.Data.RenderCSV(w, true)
}

func (f *FileResource) GetPath() string {
	return f.Path
}

func (f *FileResource) GetDescription() string {
	return f.Description
}

func (f *FileResource) WriteContent(w io.Writer) error {
	fi, err := os.Open(f.LocalPath)
	if err != nil {
		return err
	}
	defer fi.Close()
	_, err = io.Copy(w, fi)
	return err
}

func (lic License) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	fmt.Fprintf(&b, `{ "name": "%s" }`, string(lic))
	return b.Bytes(), nil
}

func (dp *DataPackage) AddResource(resouces ...Resource) {
	dp.Resources = append(dp.Resources, resouces...)
}

func (dp *DataPackage) GetMetadata() map[string]interface{} {
	var resources []interface{}
	for _, resource := range dp.Resources {
		resources = append(resources, map[string]interface{}{
			"path":        resource.GetPath(),
			"description": resource.GetDescription(),
		})
	}
	m := map[string]interface{}{
		"id":        dp.ID,
		"title":     dp.Title,
		"licenses":  dp.Licenses,
		"resources": resources,
	}
	return m
}

func (dp *DataPackage) Write(dir string) error {
	m := dp.GetMetadata()
	if err := dp.writeJSON(filepath.Join(dir, "dataset-metadata.json"), m); err != nil {
		return err
	}
	for _, resource := range dp.Resources {
		if err := dp.writeContent(dir, resource); err != nil {
			return err
		}
	}
	return nil
}

func (dp *DataPackage) writeJSON(path string, val interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(val)
}

func (dp *DataPackage) writeContent(dir string, r Resource) error {
	path := filepath.Join(dir, r.GetPath())
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, r.GetPath()))
	if err != nil {
		return err
	}
	defer f.Close()
	return r.WriteContent(f)
}
