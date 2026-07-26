package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var keyOrder = map[string]int{
	"id":           1,
	"display":      2,
	"description":  3,
	"type":         4,
	"label":        5,
	"time":         6,

	"health":       10,
	"speed":        11,
	"loot":         12,
	"range":        13,
	"hitbox":       14,
	"friendly":     15,
	"damage":       16,
	"piercing":     17,
	"airtime":      18,
	"radius":       19,
	"stats":        20,

	"w":            21,
	"h":            22,
	"x":            23,
	"y":            24,
	"angle":        25,
	"scale":        26,
	"outline":      27,
	"sprite":       28,

	"head_angle":   30,
	"body_angle":   31,
	"hand_angle":   32,
	"hand_scale":   33,
	"hand_x":       34,
	"hand_y":       35,
	"object_scale": 36,

	"animation":    40,
	"reload":       41,
	"projectiles":  42,
	"bombs":        43,
	"summons":      44,
	"wait":         45,

	"duration":     50,
	"max_health":   51,
	"min_health":   52,
	"single_use":   53,
	"priority":     54,
	"movement":     55,
	"attacks":      56,

	"modes":        60,
	"frames":       61,
	"object":       62,
	"trail":        63,
	"body":         64,
	"hand":         65,
	"equipped":     66,
	"entries":      67,

	"spawns":       76,
	"group":        77,
	"tag":          78,
	"chance":       79,
	"soulbound":    80,
	"npcs":         81,

	"children":     99,
}

func orderedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		oi, oki := keyOrder[keys[i]]
		oj, okj := keyOrder[keys[j]]
		if oki && okj {
			return oi < oj
		}
		if oki != okj {
			return oki
		}
		return keys[i] < keys[j]
	})
	return keys
}

func prettyEncode(v any, indent string) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeValue(&buf, v, indent, ""); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, v any, indent, current string) error {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 0 {
			buf.WriteString("{}")
			return nil
		}
		buf.WriteString("{\n")
		next := current + indent
		keys := orderedKeys(t)
		for i, k := range keys {
			buf.WriteString(next)
			kb, _ := json.Marshal(k)
			buf.Write(kb)
			buf.WriteString(": ")
			if err := encodeValue(buf, t[k], indent, next); err != nil {
				return err
			}
			if i < len(keys)-1 {
				buf.WriteByte(',')
			}
			buf.WriteByte('\n')
		}
		buf.WriteString(current)
		buf.WriteByte('}')
	case []any:
		if len(t) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteString("[\n")
		next := current + indent
		for i, it := range t {
			buf.WriteString(next)
			if err := encodeValue(buf, it, indent, next); err != nil {
				return err
			}
			if i < len(t)-1 {
				buf.WriteByte(',')
			}
			buf.WriteByte('\n')
		}
		buf.WriteString(current)
		buf.WriteByte(']')
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(b)
	}
	return nil
}

var (
	assetsDir string
	toolsDir  string
	clientDir string
)

var savableAssets = map[string]bool{
	"animations.json":  true,
	"npcs.json":        true,
	"projectiles.json": true,
	"bombs.json":       true,
	"items.json":       true,
	"loot.json":        true,
	"spawns.json":      true,
}

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not resolve source location")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..")
	root, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	assetsDir = filepath.Join(root, "server", "engine", "assets")
	toolsDir = filepath.Join(root, "tools")
	clientDir = filepath.Join(root, "client")
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func serveAsset(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/assets/")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	full := filepath.Join(assetsDir, name)
	http.ServeFile(w, r, full)
}

func saveAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/save/")
	if !savableAssets[name] {
		http.Error(w, "asset not savable: "+name, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	pretty, err := prettyEncode(parsed, "  ")
	if err != nil {
		http.Error(w, "marshal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	final := filepath.Join(assetsDir, name)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, pretty, 0644); err != nil {
		http.Error(w, "write tmp: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.Rename(tmp, final); err != nil {
		http.Error(w, "rename: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
	fmt.Printf("saved %s (%d bytes)\n", name, len(pretty))
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/tools/", http.StripPrefix("/tools/", http.FileServer(http.Dir(toolsDir))))
	mux.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir(clientDir))))
	mux.HandleFunc("/assets/", serveAsset)
	mux.HandleFunc("/save/", saveAsset)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/tools/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	addr := ":8083"
	fmt.Println("tools server listening on http://localhost" + addr)
	fmt.Println("  assets dir:", assetsDir)
	fmt.Println("  tools dir: ", toolsDir)
	fmt.Println("  client dir:", clientDir)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}
