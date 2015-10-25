package guestbook

import (
        "html/template"
        "net/http"
        "time"

        "appengine"
        "appengine/datastore"
        "appengine/user"
)

type Greeting struct {
        Author  string
        Content string
        Date    time.Time
}

func init() {
        http.HandleFunc("/", root)
        http.HandleFunc("/sign", sign)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
        // The string "default_guestbook" here could be varied to have multiple guestbooks.
        return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        // Ancestor queries, as shown here, are strongly consistent with the High
        // Replication Datastore. Queries that span entity groups are eventually
        // consistent. If we omitted the .Ancestor from this query there would be
        // a slight chance that Greeting that had just been written would not
        // show up in a query.
        q := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
        greetings := make([]Greeting, 0, 10)
        if _, err := q.GetAll(c, &greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        if err := guestbookTemplate.Execute(w, greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

var guestbookTemplate = template.Must(template.New("book").Parse(`
<html>
  <head>
    <title>Jimmy's white board</title>
    <link rel="shortcut icon" href="asian.ico">
  </head>
  <body>

  <center><img src="http://www.pageresource.com/wallpapers/wallpaper/cool-cat.jpg" 
      alt="cat" style="width:604px;height:428px;"/>
 <header id="top"> </center>
    <center> <h1>Hi visitors, here is a riddle for you:</h1> </center>
    <center> <h2>When is 99 more than 100?? </h2> </center>

    <nav>
     <center> <li><a href="#aboutMe">Answer</a></li> <h3>DONT CHEAT!</h3> </center>
    </nav>
 </header>

  <h2> leave your answer below </h2>
    {{range .}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Leave an answer!"></div>
    </form>
    <br>
  </body>

<br><br><br><br><br><br><br><br><br><br><br>
<section>
    <article>
        <h1><a id="aboutMe">Answer for riddle: <br> A microwave. Generally when 
        you run a microwave for ’99’ it runs for 1 minute and 39 seconds. ‘100’ 
        runs for 1 minute.</a>
        </h1>
    </article>
</section>
<br><br><br><br><br><br><br><br><br><br><br>

</html>
`))

func sign(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        g := Greeting{
                Content: r.FormValue("content"),
                Date:    time.Now(),
        }
        if u := user.Current(c); u != nil {
                g.Author = u.String()
        }
        // We set the same parent key on every Greeting entity to ensure each Greeting
        // is in the same entity group. Queries across the single entity group
        // will be consistent. However, the write rate to a single entity group
        // should be limited to ~1/second.
        key := datastore.NewIncompleteKey(c, "Greeting", guestbookKey(c))
        _, err := datastore.Put(c, key, &g)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/", http.StatusFound)
}

func static(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "public/"+r.URL.Path)
}
