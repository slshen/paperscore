package webdata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type HasPage interface {
	GetPage() *Page
}

type ResourceContent func() ([]byte, error)

type Page struct {
	ID        string
	Content   func() string
	Front     map[string]interface{}
	Pages     []HasPage
	Resources map[string]ResourceContent
}

func (p *Page) Set(name string, value interface{}) *Page {
	if p.Front == nil {
		p.Front = make(map[string]interface{})
	}
	p.Front[name] = value
	return p
}

func (p *Page) SetFrontStruct(value interface{}) *Page {
	var m map[string]interface{}
	if err := mapstructure.Decode(value, &m); err != nil {
		panic(err)
	}
	for k, v := range m {
		switch u := v.(type) {
		case string:
			p.Set(k, u)
		case int:
			p.Set(k, u)
		}
	}
	return p
}

func (p *Page) WriteFiles(dir string) error {
	pageDir := filepath.Join(dir, p.ID)
	if err := os.MkdirAll(pageDir, 0777); err != nil {
		return err
	}
	if err := p.writeIndex(pageDir); err != nil {
		return err
	}
	for name := range p.Resources {
		if err := p.writeResource(pageDir, name); err != nil {
			return err
		}
	}
	for i, c := range p.Pages {
		cp := c.GetPage()
		cp.Set("weight", i+1)
		if err := cp.WriteFiles(pageDir); err != nil {
			return err
		}
	}
	return nil
}

func (p *Page) writeIndex(pageDir string) error {
	f, err := os.Create(filepath.Join(pageDir, "_index.md"))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Printf("Writing %s\n", f.Name())
	fmt.Fprintf(f, "---\n")
	front, err := yaml.Marshal(p.Front)
	if err != nil {
		return err
	}
	_, _ = f.Write(front)
	fmt.Fprintf(f, "---\n")
	_, _ = f.WriteString(p.Content())
	return nil
}

func (p *Page) writeResource(pageDir string, name string) error {
	f, err := os.Create(filepath.Join(pageDir, name))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Printf("Writing %s\n", f.Name())
	dat, err := p.Resources[name]()
	if err != nil {
		return err
	}
	_, err = f.Write(dat)
	return err
}
