package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var replace = regexp.MustCompile(".txt")

var templates = template.Must(template.ParseFiles("tmpl/view.html", "tmpl/save.html", "tmpl/home.html"))

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	os.Mkdir("data", 0600)
	return os.WriteFile(filename, p.Body, 0600)
}
func getAllwikis() []fs.DirEntry {
	files, err := os.ReadDir("data")
	if err != nil {
		return nil
	}
	return files
}

func loadPages(title string) (*Page, error) {
	fileName := "data/" + title + ".txt"
	body, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2]) //the title is the second subexpression
	}
}
func viewHandler(w http.ResponseWriter, r *http.Request, fileTitle string) {
	p, err := loadPages(fileTitle)
	if err != nil {
		http.Redirect(w, r, "/edit/"+fileTitle, http.StatusFound)
		return
	}
	renderTemplate(w, p, "view")
}
func saveHandler(w http.ResponseWriter, r *http.Request, fileTitle string) {
	body := r.FormValue("body")
	page := Page{Title: fileTitle, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+fileTitle, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, fileTitle string) {
	p, err := loadPages(fileTitle)
	if err != nil {
		p = &Page{Title: fileTitle}
	}
	renderTemplate(w, p, "save")
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	files := getAllwikis()
	t, err := template.ParseFiles("tmpl/home.html")
	if err != nil {
		http.NotFound(w, r)
	}
	var fileNames []string
	for _, v := range files {
		fileNames = append(fileNames, replace.ReplaceAllString(v.Name(), ""))
	}
	fmt.Println(fileNames)
	t.Execute(w, fileNames)
}
func createHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("title")
	http.Redirect(w, r, "/edit/"+body, http.StatusFound)
}
func renderTemplate(w http.ResponseWriter, p *Page, file string) {
	err := templates.ExecuteTemplate(w, file+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/create", createHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
