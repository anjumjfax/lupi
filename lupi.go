/*
lupi, board software
Copyright (C) 2017-2018, Anjum Ahmed <anjumahmed at live dot co dot uk>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"io/ioutil"
	"strconv"
	"time"
)

var form string = `<html>
<form action="thread" method="POST">
<table>
<tr>
<td><label>Name</label></td>
<td><input type="text" name="name"/></td>
</tr>
<tr>
<td><label>Options</label></td>
<td><input type="text" name="email"/></td>
</tr>
<tr>
<td><label>Subject</label></td>
<td><input type="text" name="subject"/>
<input type="submit" value="post"/></td>
</tr>
<tr>
<td><label>Comment</label></td>
<td>
<textarea rows="4" cols="48" name="comment" value=""></textarea>
</td>
</tr>
</table>
</form>
</html>`

var thread_template string = `<html>
<form action="/post/{{.Post.Count}}" method="POST">
<table>
<tr>
<td><label>Name</label></td>
<td><input type="text" name="name"/></td>
</tr>
<tr>
<td><label>Options</label>
<td><input type="text" name="email"/>
<input type="submit" value="post"/></td>
</tr>
<tr>
<td><label>Comment</label></td>
<td>
<textarea rows="4" cols="48" name="comment" value=""></textarea>
</td>
</tr>
</table>
</form>
<hr/>
<div class="op">
<b>{{.Subject}}</b>
<b>{{.Post.Name}}</b>
<time datettime="{{.Post.Time}}">{{.Post.Time}}</time>
<span>No.{{.Post.Count}}</span>
<p>{{.Post.Comment}}</p>
</div>
{{$a := .Replies}}
{{range $a}}
<div class="reply">
<b>{{.Name}}</b>
<time datettime="{{.Time}}">{{.Time}}</time>
<span>No.{{.Count}}</span>
<p>{{.Comment}}</p>
</div>
{{end}}
</html>`

var count int = 0
var activeThreads []*Thread

type Post struct {
	Name    string
	Comment string
	Time    string
	Count   int
}

type Thread struct {
	Subject    string
	Post       Post
	Replies    []Post
	ReplyCount int
}

func threadFind(id int) (*Thread, bool) {
	i := 0
	for ; i < len(activeThreads) || id < len(activeThreads); i++ {
		if activeThreads[i].Post.Count == id {
			return activeThreads[i], true
		}
	}
	return nil, false
}

func threadCreate(name string, options string, subject string, comment string) {
	t := new(Thread)
	t.Replies = make([]Post, 0, 300)
	t.Subject = subject
	t.ReplyCount = 0
	t.Post = *postCreate(name, options, comment)
	fmt.Printf("New thread: %d\n", t.Post.Count)
	newlen := len(activeThreads) + 1
	activeThreads = activeThreads[:newlen]
	activeThreads[newlen-1] = t
	id_str := strconv.Itoa(t.Post.Count)
	fp, _ := os.Create(id_str)
	defer fp.Close()
	w := csv.NewWriter(fp)
	w.Write([]string{t.Post.Name, t.Post.Time, t.Post.Comment})
	w.Flush()
}

func loadCache(){
	files, err := ioutil.ReadDir("./")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, fp := range files {
		fmt.Println(fp.Name())
	}
}

func postCreate(name string, options string, comment string) *Post {
	p := new(Post)
	if name == "" {
		p.Name = "Anonymous"
	} else {
		p.Name = name
	}
	p.Comment = comment
	p.Time = time.Now().Format("01/02/06(Mon)03:04:05")
	p.Count = count
	count = count + 1
	return p
}

func postPostNew(w http.ResponseWriter, r *http.Request) {
	id_str := r.URL.Path[len("/post/"):]
	id, _ := strconv.Atoi(id_str)
	fmt.Printf("Reply thread: %d\n", id)
	thread, found := threadFind(id)

	if !found {
		http.Error(w, "could not find thread", http.StatusInternalServerError)
	}

	p := postCreate(r.FormValue("name"), r.FormValue("email"), r.FormValue("comment"))

	thread.Replies = thread.Replies[:thread.ReplyCount+1]
	thread.Replies[thread.ReplyCount] = *p
	thread.ReplyCount = thread.ReplyCount + 1
	fmt.Printf("Replying to thread no: %d", id)
	http.Redirect(w, r, "/thread/"+id_str, http.StatusFound)
}

func threadPostNew(w http.ResponseWriter, r *http.Request) {
	threadCreate(r.FormValue("name"), r.FormValue("email"), r.FormValue("subject"), r.FormValue("comment"))
	http.Redirect(w, r, "/thread/"+strconv.Itoa(count-1), http.StatusFound)
}

func threadGetShow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Path[len("/thread/"):])
	fmt.Printf("Load thread: %d", id)
	thread, found := threadFind(id)
	if !found {
		http.Error(w, "can't find thread", http.StatusNotFound)
	}
	tmpl, e := template.New("test").Parse(thread_template)
	if e != nil {
		http.Error(w, "bad template", http.StatusInternalServerError)
	}
	_ = tmpl.Execute(w, thread)
}

func boardGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(form))
}

func main() {
	activeThreads = make([]*Thread, 0, 150)
	loadCache()
	http.HandleFunc("/", boardGet)
	http.HandleFunc("/thread", threadPostNew)
	http.HandleFunc("/thread/", threadGetShow)
	http.HandleFunc("/post/", postPostNew)
	http.ListenAndServe(":8080", nil)
}
