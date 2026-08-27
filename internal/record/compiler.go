package record

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0xRech/AliveSpec/internal/observe"
	"github.com/0xRech/AliveSpec/internal/spec"
)

type processEvidence struct {
	executable   string
	observations int
}

type connectionEvidence struct {
	observations int
	processes    map[string]struct{}
}

type fileEvidence struct {
	observations int
	processes    map[string]struct{}
}

type Collector struct {
	allFiles    bool
	events      int
	processes   map[string]*processEvidence
	connections map[string]*connectionEvidence
	files       map[string]*fileEvidence
}

func NewCollector(allFiles bool) *Collector {
	return &Collector{
		allFiles:    allFiles,
		processes:   map[string]*processEvidence{},
		connections: map[string]*connectionEvidence{},
		files:       map[string]*fileEvidence{},
	}
}

func (c *Collector) Add(e observe.Event) {
	c.events++
	if e.Process != "" {
		p := c.processes[e.Process]
		if p == nil {
			p = &processEvidence{}
			c.processes[e.Process] = p
		}
		p.observations++
		if e.Kind == observe.KindProcess && e.Path != "" {
			p.executable = e.Path
		}
	}

	switch e.Kind {
	case observe.KindTCP:
		key := e.Host + ":" + itoa(e.Port)
		item := c.connections[key]
		if item == nil {
			item = &connectionEvidence{processes: map[string]struct{}{}}
			c.connections[key] = item
		}
		item.observations++
		if e.Process != "" {
			item.processes[e.Process] = struct{}{}
		}
	case observe.KindFile:
		if !c.allFiles && !interestingFile(e.Path) {
			return
		}
		item := c.files[e.Path]
		if item == nil {
			item = &fileEvidence{processes: map[string]struct{}{}}
			c.files[e.Path] = item
		}
		item.observations++
		if e.Process != "" {
			item.processes[e.Process] = struct{}{}
		}
	}
}

func (c *Collector) Contract(name string) *spec.Contract {
	host, _ := os.Hostname()
	contract := spec.New(name, host)

	processNames := sortedKeys(c.processes)
	for _, name := range processNames {
		p := c.processes[name]
		contract.Requires.Processes = append(contract.Requires.Processes, spec.ProcessRequirement{
			Name:       name,
			Executable: p.executable,
			Running:    true,
			Evidence: spec.Evidence{
				Source:       "observed",
				Observations: p.observations,
				Confidence:   0.75,
			},
		})
	}

	connectionKeys := sortedKeys(c.connections)
	for _, key := range connectionKeys {
		item := c.connections[key]
		host, port := splitEndpoint(key)
		contract.Requires.Connections = append(contract.Requires.Connections, spec.ConnectionRequirement{
			Protocol: "tcp",
			Host:     host,
			Port:     port,
			Evidence: spec.Evidence{
				Source:       "observed",
				Observations: item.observations,
				Confidence:   0.90,
				Processes:    sortedSet(item.processes),
			},
		})
	}

	filePaths := sortedKeys(c.files)
	for _, path := range filePaths {
		item := c.files[path]
		contract.Requires.Files = append(contract.Requires.Files, spec.FileRequirement{
			Path:   path,
			Exists: true,
			Evidence: spec.Evidence{
				Source:       "observed",
				Observations: item.observations,
				Confidence:   0.70,
				Processes:    sortedSet(item.processes),
			},
		})
	}

	return contract
}

func (c *Collector) Counts() (events, processes, connections, files int) {
	return c.events, len(c.processes), len(c.connections), len(c.files)
}

func interestingFile(path string) bool {
	if path == "" || !filepath.IsAbs(path) {
		return false
	}
	clean := filepath.Clean(path)
	for _, prefix := range []string{"/proc/", "/sys/", "/dev/", "/run/", "/tmp/", "/usr/lib/", "/usr/share/", "/lib/"} {
		if strings.HasPrefix(clean, prefix) {
			return false
		}
	}
	base := filepath.Base(clean)
	if base == "hosts" || base == "resolv.conf" || base == "nsswitch.conf" {
		return true
	}
	ext := strings.ToLower(filepath.Ext(clean))
	switch ext {
	case ".conf", ".config", ".ini", ".yaml", ".yml", ".json", ".toml", ".xml", ".properties", ".env", ".pem", ".crt":
		return true
	}
	lower := strings.ToLower(clean)
	return strings.Contains(lower, "/config/") || strings.Contains(lower, "/conf.d/")
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedSet(m map[string]struct{}) []string {
	return sortedKeys(m)
}

func splitEndpoint(value string) (string, int) {
	idx := strings.LastIndex(value, ":")
	if idx < 0 {
		return value, 0
	}
	return value[:idx], atoi(value[idx+1:])
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

func atoi(value string) int {
	result := 0
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return 0
		}
		result = result*10 + int(value[i]-'0')
	}
	return result
}
