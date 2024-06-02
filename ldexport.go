package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fazalmajid/go-plist"
	"rsc.io/qr"
)

var (
	IncludeArchived *bool
)

type Entry struct {
	Service  string
	Login    string
	Created  time.Time
	Modified time.Time
	URL      string
	Favorite bool
	Archived bool
}

func nstime(t float64) time.Time {
	// The NSTime epoch is 2001-01-01 UTC
	return time.Unix(int64(t)+978307200, int64(1e9*(t-float64(uint64(t)))))

}

func export(fn string, format string) {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal("could not read plist: ", err)
	}
	defer f.Close()
	p := plist.NewDecoder(f)
	dec := make(map[string][]byte, 0)
	err = p.Decode(dec)
	if err != nil {
		log.Fatal("could not decode plist: ", err)
	}
	//fmt.Println("format: ", p.Format)
	data := dec["kLDExtensionItemsKey"]
	dec2 := make(map[string]interface{}, 0)
	_, err = plist.Unmarshal(data, dec2)
	if err != nil {
		log.Fatal("could not decode nested plist: ", err)
	}
	// decode Apple's crackpot NSKeyedArchiver serialization format, see:
	// https://www.mac4n6.com/blog/2016/1/1/manual-analysis-of-nskeyedarchiver-formatted-plist-files-a-review-of-the-new-os-x-1011-recent-items
	rootid := uint64(dec2["$top"].(map[string]interface{})["root"].(plist.UID))
	objects := dec2["$objects"].([]interface{})
	root := objects[rootid].(map[string]interface{})
	all := make([]Entry, 0)
	for _, objid := range root["NS.objects"].([]interface{}) {
		key := objid.(plist.UID)
		e := objects[key].(map[string]interface{})
		entry := Entry{}
		entry.Archived = e["itemIsArchivedKey"].(bool)
		if !*IncludeArchived && entry.Archived {
			continue
		}
		entry.Favorite = e["itemFavoriteKey"].(bool)
		entry.Service = objects[e["serviceNameKey"].(plist.UID)].(string)
		entry.Login = objects[e["accountNameKey"].(plist.UID)].(string)
		entry.Created = nstime(objects[e["dateCreatedKey"].(plist.UID)].(map[string]interface{})["NS.time"].(float64))
		mod := objects[e["dateModifiedKey"].(plist.UID)]
		switch mod.(type) {
		case map[string]interface{}:
			entry.Modified = nstime(mod.(map[string]interface{})["NS.time"].(float64))
		default:
			entry.Modified = entry.Created
		}
		urlobj, ok := e["itemURLString"]
		if !ok {
			log.Fatalf("missing URL in %v\n", e)
		}
		switch urlobj.(type) {
		case plist.UID:
			entry.URL = objects[urlobj.(plist.UID)].(string)
		default:
			log.Fatalf("%s URL: %v", entry.Service, urlobj)
		}
		if !strings.HasPrefix(entry.URL, "otpauth://") {
			secret := objects[e["itemKeyKey"].(plist.UID)].(string)
			if secret == "" || secret == "$null" {
				log.Fatalf("Cannot get secret, %s URL: %v", entry.Service, e)
			}
			entry.URL = fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
				entry.Service, entry.Login, secret, entry.Service,
			)
		}
		if !ok || entry.URL == "$null" {
			log.Fatalf("missing URL in %v\n", e)
		}
		all = append(all, entry)
	}
	switch format {
	case "json":
		j, err := json.MarshalIndent(all, "", "    ")
		if err != nil {
			log.Fatalf("could not marshal %v: %v", all, err)
		}
		fmt.Println(string(j))
	case "html":
		html_export(all)
	default:
		log.Fatal("unknown export format: ", format)
	}
}

func main() {
	IncludeArchived = flag.Bool("a", false, "also include archived secrets")
	HTML := flag.Bool("html", false, "export in HTML format")
	flag.Parse()

	format := "json"
	if *HTML {
		format = "html"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("could not locate home directory: ", err)
	}
	export(home+"/Library/Containers/com.corybohon.Lockdown-Mac/Data/Library/Preferences/group.corybohon.Lockdown.plist", format)
}

const tmpl = `<!DOCTYPE html>
<html>
<head>
<title>Lockdown TOTP secrets export</title>
<style>
body {
  font-family: 'Source Sans Pro', 'Helvetica Neue Light', 'Helvetica Neue',
                Helvetica, 'Nimbus Sans L', sans-serif;
  font-size: 24px;
}
th {
  text-align: left;
  padding-right: 2em;
}
h2 {
  border-bottom: 3px solid;
  padding-bottom: 6px;
}
div.entry {
  break-inside: avoid;
}
</style>
</head>
<body>
<h1>Lockdown TOTP secrets export</h1>
<p>Generated on {{ .Now }}</p>
{{range .Entries }}
<div class="entry">
<h2>{{ .Service }}{{if (ne .Login "")}} ({{ .Login}}){{end}}</h2>
<table>
<tr><th>Service</th><td>{{ .Service }}</td></tr>
<tr><th>Login</th><td>{{ .Login }}</td></tr>
<tr>
<th>Created</th><td>{{ .Created.Format "2006-01-02 15:04:05 MST" }}</td>
</tr>
<tr>
<th>Modified</th><td>{{ .Modified.Format "2006-01-02 15:04:05 MST" }}</td>
</tr>
<tr><th>URL</th><td>{{ .URL }}</td></tr>
<tr>
<th>Favorite</th><td>{{if .Favorite}}&#x2705{{else}}&#x274C{{end}}</td>
</tr>
<tr>
<th>Archived</th><td>{{if .Archived}}&#x2705{{else}}&#x274C{{end}}</td>
</tr>
<tr>
<th>QR code</qr><td><img class="QR" src="{{ .QR }}" alt="{{ .URL }}"></td>
</tr>
</table>
</div>
{{end}}
</body>
</html>
`

func (e *Entry) QR() template.URL {
	code, err := qr.Encode(e.URL, qr.Q) // 55% redundant error-correction level
	if err != nil {
		log.Fatal("could not QR-encode ", e.URL, ": ", err)
	}
	b64 := base64.StdEncoding.EncodeToString(code.PNG())
	enc := "data:image/png;base64," + b64
	return template.URL(enc)
}

func html_export(all []Entry) {
	t := template.New("Lockdown export")
	var err error
	t, err = t.Parse(tmpl)
	if err != nil {
		log.Fatal("could not parse embedded template ", tmpl, ": ", err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"Now":     time.Now().Format("2006-01-02 15:04:05 MST"),
		"Entries": all,
	})
	if err != nil {
		log.Fatal("error rendering HTML", err)
	}
	_, err = os.Stdout.Write(buf.Bytes())
	if err != nil {
		log.Fatal("error writing HTML: ", err)
	}
}
