package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var (
	startTime    = time.Now()
	requestCount uint64
)

type EnvItem struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsCustom bool   `json:"isCustom"`
}

type DiagnosticsData struct {
	AppVersion     string    `json:"appVersion"`
	Hostname       string    `json:"hostname"`
	HostHeader     string    `json:"hostHeader"`
	Proto          string    `json:"proto"`
	RemoteAddr     string    `json:"remoteAddr"`
	ForwardedFor   string    `json:"forwardedFor"`
	ForwardedHost  string    `json:"forwardedHost"`
	Uptime         string    `json:"uptime"`
	UptimeSeconds  int64     `json:"uptimeSeconds"`
	GoVersion      string    `json:"goVersion"`
	OS             string    `json:"os"`
	Arch           string    `json:"arch"`
	NumCPU         int       `json:"numCpu"`
	Goroutines     int       `json:"goroutines"`
	AllocMemMB     float64   `json:"allocMemMb"`
	TotalRequests  uint64    `json:"totalRequests"`
	CurrentTime    string    `json:"currentTime"`
	EnvVariables   []EnvItem `json:"envVariables"`
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Takō Cloud Verification Mini App</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap');
    body {
      font-family: 'Plus Jakarta Sans', sans-serif;
      background-color: #08090E;
      color: #F8FAFC;
    }
    .font-mono {
      font-family: 'JetBrains Mono', monospace;
    }
  </style>
</head>
<body class="min-h-screen p-4 sm:p-8 selection:bg-indigo-500/30">
  <div class="max-w-5xl mx-auto space-y-6">
    
    <!-- Top Bar -->
    <header class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-white/10">
      <div class="space-y-1">
        <div class="flex items-center gap-2.5">
          <div class="size-9 rounded-xl bg-gradient-to-tr from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/20 text-white font-bold text-lg">
            🐙
          </div>
          <div>
            <h1 class="text-xl font-bold tracking-tight text-white flex items-center gap-2">
              Takō Verification Mini App
              <span class="inline-flex items-center gap-1 text-[11px] font-semibold px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 font-mono">
                <span class="size-1.5 rounded-full bg-emerald-400 animate-pulse"></span> {{.AppVersion}}
              </span>
            </h1>
            <p class="text-xs text-slate-400">Validating Ingress Domains, Environment Variables, & Container Execution</p>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 text-xs font-mono">
        <div class="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-slate-300">
          Uptime: <span class="text-indigo-400 font-semibold">{{.Uptime}}</span>
        </div>
        <div class="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-slate-300">
          Reqs: <span class="text-indigo-400 font-semibold">{{.TotalRequests}}</span>
        </div>
      </div>
    </header>

    <!-- 1. Domain & Ingress Routing Card -->
    <section class="rounded-2xl border border-white/10 bg-[#0E1017] p-5 sm:p-6 space-y-4">
      <div class="flex items-center justify-between border-b border-white/5 pb-3">
        <div class="flex items-center gap-2">
          <span class="text-lg">🌐</span>
          <h2 class="text-sm font-bold uppercase tracking-wider text-slate-200 font-mono">1. Domain & Ingress Routing Verification</h2>
        </div>
        <span class="text-xs font-mono text-indigo-400 bg-indigo-500/10 px-2 py-0.5 rounded border border-indigo-500/20">
          Traefik Edge
        </span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs font-mono">
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Requested Host</div>
          <div class="text-white font-bold text-sm mt-0.5 text-indigo-400 truncate" title="{{.HostHeader}}">{{.HostHeader}}</div>
        </div>
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Forwarded Proto (SSL)</div>
          <div class="text-emerald-400 font-bold text-sm mt-0.5 uppercase">{{if .ForwardedHost}}{{.Proto}}{{else}}HTTP / DIRECT{{end}}</div>
        </div>
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Forwarded Host</div>
          <div class="text-slate-200 font-semibold text-sm mt-0.5 truncate">{{if .ForwardedHost}}{{.ForwardedHost}}{{else}}None (Direct){{end}}</div>
        </div>
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Client IP</div>
          <div class="text-slate-200 font-semibold text-sm mt-0.5 truncate">{{if .ForwardedFor}}{{.ForwardedFor}}{{else}}{{.RemoteAddr}}{{end}}</div>
        </div>
      </div>

      <div class="p-3 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-xs text-indigo-300 flex items-center justify-between">
        <span>Target Custom Domain: <strong class="text-white font-mono">demo.gettako.dev</strong></span>
        <span class="font-mono text-[11px] text-indigo-200">DNS & Ingress Verified</span>
      </div>
    </section>

    <!-- 2. Injected Environment Variables Card -->
    <section class="rounded-2xl border border-white/10 bg-[#0E1017] p-5 sm:p-6 space-y-4">
      <div class="flex items-center justify-between border-b border-white/5 pb-3">
        <div class="flex items-center gap-2">
          <span class="text-lg">🔑</span>
          <h2 class="text-sm font-bold uppercase tracking-wider text-slate-200 font-mono">2. Environment Variables Verification</h2>
        </div>
        <span class="text-xs font-mono text-slate-400">{{len .EnvVariables}} Variables Loaded</span>
      </div>

      <div class="overflow-x-auto rounded-xl border border-white/5">
        <table class="w-full text-left text-xs font-mono">
          <thead class="bg-white/[0.04] text-slate-400 border-b border-white/5">
            <tr>
              <th class="px-4 py-2.5">Key</th>
              <th class="px-4 py-2.5">Value</th>
              <th class="px-4 py-2.5 text-right">Scope</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            {{range .EnvVariables}}
            <tr class="hover:bg-white/[0.02] transition-colors {{if .IsCustom}}bg-indigo-500/[0.03]{{end}}">
              <td class="px-4 py-2.5 font-bold {{if .IsCustom}}text-indigo-400{{else}}text-slate-300{{end}}">
                {{.Key}}
              </td>
              <td class="px-4 py-2.5 text-slate-300 break-all">
                {{.Value}}
              </td>
              <td class="px-4 py-2.5 text-right">
                {{if .IsCustom}}
                <span class="px-2 py-0.5 rounded bg-indigo-500/20 border border-indigo-500/30 text-indigo-300 text-[10px]">
                  Custom Injected
                </span>
                {{else}}
                <span class="px-2 py-0.5 rounded bg-white/5 text-slate-400 text-[10px]">
                  Runtime System
                </span>
                {{end}}
              </td>
            </tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </section>

    <!-- 3. Container & Host Diagnostics -->
    <section class="rounded-2xl border border-white/10 bg-[#0E1017] p-5 sm:p-6 space-y-4">
      <div class="flex items-center justify-between border-b border-white/5 pb-3">
        <div class="flex items-center gap-2">
          <span class="text-lg">💻</span>
          <h2 class="text-sm font-bold uppercase tracking-wider text-slate-200 font-mono">3. Container & Terminal Runtime Spec</h2>
        </div>
        <span class="text-xs font-mono text-slate-400">{{.OS}} / {{.Arch}}</span>
      </div>

      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Container Hostname</div>
          <div class="text-white font-semibold mt-0.5 truncate">{{.Hostname}}</div>
        </div>
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Go Runtime</div>
          <div class="text-slate-200 font-semibold mt-0.5 truncate">{{.GoVersion}}</div>
        </div>
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">Allocated Memory</div>
          <div class="text-slate-200 font-semibold mt-0.5">{{printf "%.2f" .AllocMemMB}} MB</div>
        </div>
        <div class="p-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div class="text-slate-400 text-[11px]">CPU Cores / Goroutines</div>
          <div class="text-slate-200 font-semibold mt-0.5">{{.NumCPU}} Cores / {{.Goroutines}}</div>
        </div>
      </div>

      <div class="p-3.5 rounded-xl bg-white/[0.02] border border-white/5 text-xs text-slate-400 space-y-1">
        <div class="font-bold text-slate-300 font-mono">Terminal Exec Diagnostics:</div>
        <p>This container runs an Alpine Linux base with full shell tools (<code class="text-indigo-400">sh</code>, <code class="text-indigo-400">curl</code>, <code class="text-indigo-400">cat</code>, <code class="text-indigo-400">ps</code>, <code class="text-indigo-400">top</code>, <code class="text-indigo-400">env</code>). Open the <strong class="text-white">Terminal</strong> tab in Takō Console to execute live interactive commands directly.</p>
      </div>
    </section>

    <!-- Footer Links -->
    <footer class="pt-4 pb-8 flex items-center justify-between text-xs font-mono text-slate-500 border-t border-white/5">
      <div>Deployed on <a href="https://gettako.dev" target="_blank" class="text-indigo-400 hover:underline">Takō Cloud PaaS</a></div>
      <div class="flex items-center gap-3">
        <a href="/api/info" target="_blank" class="hover:text-slate-300">/api/info</a>
        <span>•</span>
        <a href="/api/env" target="_blank" class="hover:text-slate-300">/api/env</a>
        <span>•</span>
        <a href="/health" target="_blank" class="hover:text-slate-300">/health</a>
      </div>
    </footer>
  </div>
</body>
</html>`

func getDiagnostics(r *http.Request) DiagnosticsData {
	atomic.AddUint64(&requestCount, 1)

	hostname, _ := os.Hostname()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	// Collect environment variables
	rawEnv := os.Environ()
	var envList []EnvItem
	for _, e := range rawEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			k := parts[0]
			v := parts[1]

			// Mask sensitive credentials
			lower := strings.ToLower(k)
			if strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "key") {
				if len(v) > 4 {
					v = v[:2] + "••••••••" + v[len(v)-2:]
				} else {
					v = "••••••••"
				}
			}

			isCustom := !strings.HasPrefix(k, "PATH") &&
				!strings.HasPrefix(k, "HOSTNAME") &&
				!strings.HasPrefix(k, "HOME") &&
				!strings.HasPrefix(k, "SHLVL") &&
				!strings.HasPrefix(k, "PWD")

			envList = append(envList, EnvItem{
				Key:      k,
				Value:    v,
				IsCustom: isCustom,
			})
		}
	}

	sort.Slice(envList, func(i, j int) bool {
		if envList[i].IsCustom != envList[j].IsCustom {
			return envList[i].IsCustom
		}
		return envList[i].Key < envList[j].Key
	})

	uptimeDur := time.Since(startTime)
	uptimeStr := fmt.Sprintf("%02dh %02dm %02ds",
		int(uptimeDur.Hours()),
		int(uptimeDur.Minutes())%60,
		int(uptimeDur.Seconds())%60,
	)

	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	return DiagnosticsData{
		AppVersion:    "v1.1.0",
		Hostname:      hostname,
		HostHeader:    r.Host,
		Proto:         proto,
		RemoteAddr:    r.RemoteAddr,
		ForwardedFor:  r.Header.Get("X-Forwarded-For"),
		ForwardedHost: r.Header.Get("X-Forwarded-Host"),
		Uptime:        uptimeStr,
		UptimeSeconds: int64(uptimeDur.Seconds()),
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		Goroutines:    runtime.NumGoroutine(),
		AllocMemMB:    float64(mem.Alloc) / 1024 / 1024,
		TotalRequests: atomic.LoadUint64(&requestCount),
		CurrentTime:   time.Now().UTC().Format(time.RFC3339),
		EnvVariables:  envList,
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Template parse error: %v", err)
	}

	mux := http.NewServeMux()

	// 1. Dashboard View
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data := getDiagnostics(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, data)
	})

	// 2. Health Endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"uptime": time.Since(startTime).String(),
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 3. API Info Endpoint
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		data := getDiagnostics(r)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	})

	// 4. API Env Endpoint
	mux.HandleFunc("/api/env", func(w http.ResponseWriter, r *http.Request) {
		data := getDiagnostics(r)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data.EnvVariables)
	})

	// 5. Echo Test Endpoint
	mux.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("msg")
		if msg == "" {
			msg = "Hello from Takō Cloud!"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"echo": msg,
			"host": r.Host,
			"time": time.Now().UTC().Format(time.RFC3339),
		})
	})

	serverAddr := ":" + port
	log.Printf("🐙 Takō Verification Mini App starting on %s...", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}
