//go:build !solution

package httpgauge

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Gauge struct {
	mu         sync.Mutex
	statistics map[string]int
}

func New() *Gauge {
	return &Gauge{
		statistics: make(map[string]int),
	}
}

func (g *Gauge) Snapshot() map[string]int {
	g.mu.Lock()
	defer g.mu.Unlock()

	snapshot := make(map[string]int, len(g.statistics))
	for pattern, count := range g.statistics {
		snapshot[pattern] = count
	}
	return snapshot
}

func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snapshot := g.Snapshot()

	patterns := make([]string, 0, len(snapshot))
	for pattern := range snapshot {
		patterns = append(patterns, pattern)
	}
	sort.Strings(patterns)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	for _, pattern := range patterns {
		_, _ = fmt.Fprintf(w, "%s %d\n", pattern, snapshot[pattern])
	}
}

func patternFromPath(path string) string {
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	parts := strings.Split(path, "/")
	patternParts := make([]string, len(parts))
	copy(patternParts, parts)

	for i, part := range parts {
		if _, err := strconv.Atoi(part); err == nil {
			if i > 0 && parts[i-1] != "" {
				patternParts[i] = "{" + parts[i-1] + "ID}"
			} else {
				patternParts[i] = "{ID}"
			}
		}
	}
	return strings.Join(patternParts, "/")
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pattern := patternFromPath(path)

		g.mu.Lock()
		g.statistics[pattern]++
		g.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
