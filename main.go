package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

const (
	VERSION = "1.0.0"
	BANNER  = `
   ___              ___           _           ___                  _
  / _ \___         /   \___  _ __| | __      / _ \___   ___   __ _| | ___
 / /_\/ _ \ _____ / /\ / _ \| '__| |/ /____ / /_\/ _ \ / _ \ / _" | |/ _ \
/ /_\\ (_) |_____/ /_// (_) | |  |   <_____/ /_\\ (_) | (_) | (_| | |  __/
\____/\___/     /___,' \___/|_|  |_|\_\    \____/\___/ \___/ \__, |_|\___|
                                                             |___/        
    Advanced Google Dorking Tool v%s
    `
)

type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
	TRACE
)

type Logger struct {
	*log.Logger
	level LogLevel
}

type Result struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Snippet    string   `json:"snippet"`
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains,omitempty"`
}

type Config struct {
	GoogleAPI   []string `yaml:"Google-API"`
	GoogleCSEID []string `yaml:"Google-CSE-ID"`
}

type SubdomainSet struct {
	items map[string]struct{}
	mu    sync.RWMutex
}

type SearchResult struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Error      string   `json:"error,omitempty"`
}

var (
	queryArg     = flag.String("q", "", "Google dorking query for your target")
	domainArg    = flag.String("d", "", "Target name for Google dorking")
	outputArg    = flag.String("o", "", "File name to save the dorking results")
	formatArg    = flag.String("format", "txt", "Output format (txt, json, csv)")
	subdomains   = flag.Bool("subs", false, "Only output found subdomains")
	concurrent   = flag.Int("concurrent", 10, "Number of concurrent searches")
	verbosity    = flag.Int("v", 1, "Verbosity level (0=ERROR, 1=INFO, 2=DEBUG, 3=TRACE)")
	showVersion  = flag.Bool("version", false, "Show version information")
	noColor      = flag.Bool("no-color", false, "Disable color output")
	silent       = flag.Bool("silent", false, "Silent mode - only output results")
	timeout      = flag.Duration("timeout", 5*time.Minute, "Timeout for the entire search operation")
	results      []Result
	resultsMutex sync.Mutex
	subdomainSet = NewSubdomainSet()
	logger       *Logger
)

var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

func NewSubdomainSet() *SubdomainSet {
	return &SubdomainSet{
		items: make(map[string]struct{}),
	}
}

func (s *SubdomainSet) Add(subdomain string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[subdomain] = struct{}{}
}

func (s *SubdomainSet) ToSlice() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level >= ERROR && !*silent {
		l.Printf("%s[ERROR]%s "+format, append([]interface{}{colorRed, colorReset}, v...)...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level >= INFO && !*silent {
		l.Printf("%s[INFO]%s "+format, append([]interface{}{colorBlue, colorReset}, v...)...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level >= DEBUG && !*silent {
		l.Printf("%s[DEBUG]%s "+format, append([]interface{}{colorYellow, colorReset}, v...)...)
	}
}

func (l *Logger) Trace(format string, v ...interface{}) {
	if l.level >= TRACE && !*silent {
		l.Printf("%s[TRACE]%s "+format, append([]interface{}{colorPurple, colorReset}, v...)...)
	}
}

func init() {
	flag.Usage = func() {
		fmt.Printf(BANNER, VERSION)
		fmt.Println("\nUsage:")
		fmt.Println("  google-dorker [options] [additional-domains...]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  google-dorker -d example.com -subs -format json")
		fmt.Println("  google-dorker -d example.com -subs -silent")
		fmt.Println("  google-dorker -d example.com -concurrent 20 -format csv -o results.csv")
		fmt.Println("  google-dorker -d example.com sub1.example.com sub2.example.com -subs\n")
	}
}

func setupLogger() {
	if *noColor {
		colorReset = ""
		colorRed = ""
		colorGreen = ""
		colorYellow = ""
		colorBlue = ""
		colorPurple = ""
		colorCyan = ""
	}

	logger = &Logger{
		Logger: log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		level:  LogLevel(*verbosity),
	}
}

func loadConfig() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory: %v", err)
		os.Exit(1)
	}

	configLocations := []string{
		"google_dorker.yaml",
		filepath.Join(homeDir, ".config/google_dorker.yaml"),
		"/etc/google_dorker.yaml",
	}

	var configPath string
	for _, loc := range configLocations {
		if _, err := os.Stat(loc); err == nil {
			configPath = loc
			break
		}
	}

	if configPath == "" {
		logger.Error("Config file not found. Checked locations:")
		for _, loc := range configLocations {
			logger.Error("- %s", loc)
		}
		os.Exit(1)
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		logger.Error("Failed to get absolute path: %v", err)
		os.Exit(1)
	}

	logger.Debug("Loading configuration from: %s", absPath)
	return absPath
}

func loadAPIConfig(filename string) Config {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error("Failed to read config file: %v", err)
		os.Exit(1)
	}

	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		logger.Error("Failed to parse config file: %v", err)
		os.Exit(1)
	}

	if len(config.GoogleAPI) == 0 || len(config.GoogleCSEID) == 0 {
		logger.Error("Google API key or CSE ID missing from config")
		os.Exit(1)
	}

	return config
}

func extractSubdomains(domain, urlStr string) []string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Debug("Failed to parse URL %s: %v", urlStr, err)
		return nil
	}

	host := parsedURL.Hostname()
	if !strings.HasSuffix(host, domain) {
		return nil
	}

	if host != domain {
		subdomainSet.Add(host)
		logger.Debug("Found subdomain: %s", host)
	}
	return subdomainSet.ToSlice()
}

func constructQuery(domain, query string) string {
	if query != "" && domain != "" {
		return fmt.Sprintf("site:%s %s", domain, query)
	} else if query != "" {
		return query
	}
	return fmt.Sprintf("site:%s", domain)
}

func performSearch(ctx context.Context, svc *customsearch.Service, cseID, query string, domain string, results chan<- SearchResult) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered from panic in search routine: %v", r)
			results <- SearchResult{
				Domain: domain,
				Error:  fmt.Sprintf("Search routine panic: %v", r),
			}
		}
	}()

	localSet := NewSubdomainSet()
	startIndex := int64(1)
	totalResults := int64(100)
	resultsPerPage := int64(10)

	for startIndex < totalResults {
		select {
		case <-ctx.Done():
			results <- SearchResult{
				Domain: domain,
				Error:  "Search timeout",
			}
			return
		default:
			logger.Trace("Searching page starting at index: %d for domain: %s", startIndex, domain)
			req := svc.Cse.List().Cx(cseID).Q(query).Num(resultsPerPage).Start(startIndex)
			resp, err := req.Do()
			if err != nil {
				logger.Error("Search failed for domain %s: %v", domain, err)
				results <- SearchResult{
					Domain: domain,
					Error:  fmt.Sprintf("Search failed: %v", err),
				}
				return
			}

			if resp.Items == nil {
				break
			}

			for _, item := range resp.Items {
				if *subdomains {
					if subs := extractSubdomains(domain, item.Link); len(subs) > 0 {
						for _, sub := range subs {
							localSet.Add(sub)
						}
					}
				}
				logger.Info("%sFound:%s %s", colorGreen, colorReset, item.Link)
			}

			startIndex += resultsPerPage
			if len(resp.Items) < int(resultsPerPage) {
				break
			}

			time.Sleep(time.Second) // Rate limiting
		}
	}

	results <- SearchResult{
		Domain:     domain,
		Subdomains: localSet.ToSlice(),
	}
}

func processDomains(domains []string, svc *customsearch.Service, cseID string) map[string][]string {
	resultsChan := make(chan SearchResult, len(domains))
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var wg sync.WaitGroup
	sem := make(chan bool, *concurrent)

	for _, domain := range domains {
		logger.Info("Starting search for domain: %s", domain)
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			sem <- true
			performSearch(ctx, svc, cseID, constructQuery(d, *queryArg), d, resultsChan)
			<-sem
		}(domain)
	}

	wg.Wait()
	close(resultsChan)

	results := make(map[string][]string)
	for result := range resultsChan {
		if result.Error != "" {
			logger.Error("Error for domain %s: %s", result.Domain, result.Error)
		} else {
			results[result.Domain] = result.Subdomains
		}
	}
	return results
}

func getAllDomains() []string {
	domains := []string{*domainArg}
	domains = append(domains, flag.Args()...) // Add any additional domains from command line args
	return domains
}

func outputSubdomains(results map[string][]string) {
	switch *formatArg {
	case "json":
		if err := outputJSON(results); err != nil && !*silent {
			logger.Error("Failed to output JSON: %v", err)
		}
	case "txt":
		outputTXT(results)
	case "csv":
		if err := outputCSV(results); err != nil && !*silent {
			logger.Error("Failed to output CSV: %v", err)
		}
	default:
		outputTXT(results)
	}
}

func outputJSON(results map[string][]string) error {
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	if *outputArg != "" {
		return ioutil.WriteFile(*outputArg, output, 0644)
	}
	fmt.Println(string(output))
	return nil
}

func outputTXT(results map[string][]string) {
	var output strings.Builder
	for domain, subdomains := range results {
		if len(results) > 1 {
			output.WriteString(fmt.Sprintf("%s:\n", domain))
		}
		for _, subdomain := range subdomains {
			output.WriteString(subdomain + "\n")
		}
		if len(results) > 1 {
			output.WriteString("\n")
		}
	}

	if *outputArg != "" {
		ioutil.WriteFile(*outputArg, []byte(output.String()), 0644)
		return
	}
	fmt.Print(output.String())
}

func outputCSV(results map[string][]string) error {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	writer.Write([]string{"Domain", "Subdomain"})

	for domain, subdomains := range results {
		for _, subdomain := range subdomains {
			writer.Write([]string{domain, subdomain})
		}
	}
	writer.Flush()

	if *outputArg != "" {
		return ioutil.WriteFile(*outputArg, []byte(output.String()), 0644)
	}
	fmt.Print(output.String())
	return nil
}

func main() {
	startTime := time.Now()
	flag.Parse()

	if *showVersion && !*silent {
		fmt.Printf(BANNER, VERSION)
		return
	}

	setupLogger()
	if !*silent {
		logger.Info("Starting Google Dorker v%s", VERSION)
	}

	if *domainArg == "" {
		if !*silent {
			flag.Usage()
		}
		os.Exit(1)
	}

	configFile := loadConfig()
	config := loadAPIConfig(configFile)
	logger.Debug("Configuration loaded successfully")

	rand.Seed(time.Now().UnixNano())
	googleAPI := config.GoogleAPI[rand.Intn(len(config.GoogleAPI))]
	googleCSEID := config.GoogleCSEID[rand.Intn(len(config.GoogleCSEID))]

	ctx := context.Background()
	svc, err := customsearch.NewService(ctx, option.WithAPIKey(googleAPI))
	if err != nil {
		logger.Error("Failed to create custom search service: %v", err)
		os.Exit(1)
	}

	domains := getAllDomains()
	results := processDomains(domains, svc, googleCSEID)

	if *subdomains {
		outputSubdomains(results)
	}

	if !*silent && !*subdomains {
		duration := time.Since(startTime)
		logger.Info("%sExecution time: %v%s", colorCyan, duration, colorReset)
	}
}
