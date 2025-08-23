package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	path := flag.String("p", "", "dot path to value (use indexes for arrays, e.g., component.measures.0.value)")
	flag.Parse()
	if *path == "" {
		fmt.Fprintln(os.Stderr, "jsonval: missing -p path")
		os.Exit(2)
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jsonval: read stdin:", err)
		os.Exit(1)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintln(os.Stderr, "jsonval: parse:", err)
		os.Exit(1)
	}
	val, ok := traverse(v, strings.Split(*path, "."))
	if !ok {
		os.Exit(3)
	}
	switch t := val.(type) {
	case string:
		fmt.Print(t)
	case float64:
		// print without trailing .0 when integer-like
		if t == float64(int64(t)) {
			fmt.Printf("%d", int64(t))
		} else {
			fmt.Printf("%g", t)
		}
	case bool:
		if t {
			fmt.Print("true")
		} else {
			fmt.Print("false")
		}
	default:
		b, _ := json.Marshal(t)
		fmt.Print(string(b))
	}
}

func traverse(v any, parts []string) (any, bool) {
	cur := v
	for _, p := range parts {
		if m, ok := cur.(map[string]any); ok {
			nv, ok := m[p]
			if !ok {
				return nil, false
			}
			cur = nv
			continue
		}
		if arr, ok := cur.([]any); ok {
			// p expected to be an index
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(arr) {
				return nil, false
			}
			cur = arr[idx]
			continue
		}
		return nil, false
	}
	return cur, true
}
