package cbuildapp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

type compileCommandEntry struct {
	Directory string   `json:"directory"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
	File      string   `json:"file"`
}

type headerImpactEntry struct {
	Path             string
	TUs              map[string]struct{}
	TUCount          int
	Lines            int
	ClosureLines     int
	Impact           int
	DirectImpact     int
	TotalImpact      int
	SortedTUs        []string
	SampleCompileCmd *compileCommandResolved
}

type compileCommandResolved struct {
	Compiler     string
	Args         []string
	Directory    string
	SourceArgIdx int
}

type tuWorkItem struct {
	DBPath string
	Entry  compileCommandEntry
}

type tuResult struct {
	DBPath     string
	TUPath     string
	Resolved   *compileCommandResolved
	Transitive []string
	Err        error
}

type tuImpactEntry struct {
	TUPath            string
	PreprocessedLines int
	IncludedFiles     int
}

type headerCycleEntry struct {
	Headers []string
}

func runInspectHeaders(ctx context.Context, args []string) error {
	workspacePath := "."
	if workspacePathRaw := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter); workspacePathRaw != nil && *workspacePathRaw != "" {
		workspacePath = *workspacePathRaw
	}

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchainNames, err := inspectToolchains(ctx, ws)
	if err != nil {
		return err
	}
	if len(toolchainNames) == 0 {
		return fmt.Errorf("no toolchains found")
	}

	configs := inspectConfigs(ctx, ws)
	selectedTargets, hasTargetFilter := inspectTargets(ctx, ws)
	tuTargets, headerRoots, err := resolveInspectScope(ctx, ws, selectedTargets, hasTargetFilter)
	if err != nil {
		return err
	}

	sourceRoots, err := inspectSourceRoots(ws)
	if err != nil {
		return err
	}

	dbPaths, missingPaths, err := inspectCompileCommandsPaths(ctx, ws, toolchainNames, configs, tuTargets)
	if err != nil {
		return err
	}
	if len(dbPaths) == 0 {
		return fmt.Errorf("no compile_commands.json files found for selected targets/toolchains/configs; run cbuild build first")
	}
	if len(missingPaths) > 0 {
		fmt.Printf("Skipping %d build directories without compile_commands.json\n", len(missingPaths))
	}

	workItems := make([]tuWorkItem, 0)
	for _, dbPath := range dbPaths {
		entries, err := readCompileCommands(dbPath)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", dbPath, err)
		}
		for _, entry := range entries {
			workItems = append(workItems, tuWorkItem{
				DBPath: dbPath,
				Entry:  entry,
			})
		}
	}

	tuResults, err := analyzeTranslationUnits(ctx, workItems)
	if err != nil {
		return err
	}

	stats := make(map[string]*headerImpactEntry)
	for _, res := range tuResults {
		seenInTU := make(map[string]struct{})
		for _, dep := range res.Transitive {
			depAbs, err := absolutePath(res.Resolved.Directory, dep)
			if err != nil {
				continue
			}
			if depAbs == res.TUPath {
				continue
			}
			if _, seen := seenInTU[depAbs]; seen {
				continue
			}
			if !pathInAnyRoot(depAbs, sourceRoots) {
				continue
			}
			if !pathInAnyRoot(depAbs, headerRoots) {
				continue
			}

			seenInTU[depAbs] = struct{}{}
			item := ensureHeaderEntry(stats, depAbs)
			item.TUs[res.TUPath] = struct{}{}
			if item.SampleCompileCmd == nil {
				item.SampleCompileCmd = cloneResolvedCompileCommand(res.Resolved)
			}
		}

	}

	report := make([]headerImpactEntry, 0, len(stats))
	lineCache := make(map[string]int)
	var lineCacheMu sync.Mutex
	for _, item := range stats {
		lines, err := countLines(item.Path)
		if err != nil {
			return fmt.Errorf("failed counting lines for %s: %w", item.Path, err)
		}
		item.Lines = lines
		item.TUCount = len(item.TUs)
		item.Impact = item.Lines * item.TUCount
		item.DirectImpact = item.Lines * item.TUCount
		item.ClosureLines = item.Lines
		item.TotalImpact = item.ClosureLines * item.TUCount
		item.SortedTUs = mapKeysSorted(item.TUs)
		report = append(report, *item)
	}

	closureDeps, err := fillClosureLinesParallel(ctx, report, sourceRoots, lineCache, &lineCacheMu)
	if err != nil {
		return err
	}

	cycles := detectHeaderCycles(report, closureDeps)
	topCycles := limitHeaderCycles(cycles, 50)

	tuImpacts, err := computeTUImpacts(tuResults, sourceRoots, lineCache, &lineCacheMu)
	if err != nil {
		return err
	}

	sortHeaderImpact(report)
	directSorted := append([]headerImpactEntry{}, report...)
	sort.Slice(directSorted, func(i, j int) bool {
		if directSorted[i].DirectImpact != directSorted[j].DirectImpact {
			return directSorted[i].DirectImpact > directSorted[j].DirectImpact
		}
		if directSorted[i].TUCount != directSorted[j].TUCount {
			return directSorted[i].TUCount > directSorted[j].TUCount
		}
		if directSorted[i].Lines != directSorted[j].Lines {
			return directSorted[i].Lines > directSorted[j].Lines
		}
		return directSorted[i].Path < directSorted[j].Path
	})
	totalSorted := append([]headerImpactEntry{}, report...)
	sort.Slice(totalSorted, func(i, j int) bool {
		if totalSorted[i].TotalImpact != totalSorted[j].TotalImpact {
			return totalSorted[i].TotalImpact > totalSorted[j].TotalImpact
		}
		if totalSorted[i].ClosureLines != totalSorted[j].ClosureLines {
			return totalSorted[i].ClosureLines > totalSorted[j].ClosureLines
		}
		if totalSorted[i].TUCount != totalSorted[j].TUCount {
			return totalSorted[i].TUCount > totalSorted[j].TUCount
		}
		return totalSorted[i].Path < totalSorted[j].Path
	})

	closureSorted := append([]headerImpactEntry{}, report...)
	sort.Slice(closureSorted, func(i, j int) bool {
		if closureSorted[i].ClosureLines != closureSorted[j].ClosureLines {
			return closureSorted[i].ClosureLines > closureSorted[j].ClosureLines
		}
		if closureSorted[i].TUCount != closureSorted[j].TUCount {
			return closureSorted[i].TUCount > closureSorted[j].TUCount
		}
		return closureSorted[i].Path < closureSorted[j].Path
	})

	if len(report) == 0 {
		fmt.Println("No project headers found in compile dependencies.")
		return nil
	}

	topN := 50
	if len(directSorted) < topN {
		topN = len(directSorted)
	}
	topDirect := directSorted[:topN]
	topTotalN := 50
	if len(totalSorted) < topTotalN {
		topTotalN = len(totalSorted)
	}
	topTotal := totalSorted[:topTotalN]
	topClosureN := 50
	if len(closureSorted) < topClosureN {
		topClosureN = len(closureSorted)
	}
	topClosure := closureSorted[:topClosureN]

	reportPath, err := writeHeaderImpactReport(workspacePath, topTotal, topDirect, topClosure, report, tuImpacts, topCycles)
	if err != nil {
		return err
	}

	consoleN := 10
	if len(topTotal) < consoleN {
		consoleN = len(topTotal)
	}
	fmt.Printf("Top %d headers by total_impact\n", consoleN)
	fmt.Printf("%-14s %-14s %-13s %-10s %s\n", "direct_impact", "total_impact", "included_tus", "lines", "header")
	for _, item := range topTotal[:consoleN] {
		fmt.Printf("%-14d %-14d %-13d %-10d %s\n", item.DirectImpact, item.TotalImpact, item.TUCount, item.Lines, item.Path)
	}

	consoleDirectN := 10
	if len(topDirect) < consoleDirectN {
		consoleDirectN = len(topDirect)
	}
	fmt.Printf("\nTop %d headers by direct_impact\n", consoleDirectN)
	fmt.Printf("%-14s %-14s %-13s %-10s %s\n", "direct_impact", "total_impact", "included_tus", "lines", "header")
	for _, item := range topDirect[:consoleDirectN] {
		fmt.Printf("%-14d %-14d %-13d %-10d %s\n", item.DirectImpact, item.TotalImpact, item.TUCount, item.Lines, item.Path)
	}

	consoleTUN := 10
	if len(tuImpacts) < consoleTUN {
		consoleTUN = len(tuImpacts)
	}
	fmt.Printf("\nTop %d TUs by preprocessed lines\n", consoleTUN)
	fmt.Printf("%-18s %-14s %s\n", "preprocessed_lines", "included_files", "translation_unit")
	for _, item := range tuImpacts[:consoleTUN] {
		fmt.Printf("%-18d %-14d %s\n", item.PreprocessedLines, item.IncludedFiles, item.TUPath)
	}

	consoleCycleN := 10
	if len(topCycles) < consoleCycleN {
		consoleCycleN = len(topCycles)
	}
	fmt.Printf("\nTop %d header dependency cycles\n", consoleCycleN)
	fmt.Printf("%-8s %s\n", "size", "headers")
	if consoleCycleN == 0 {
		fmt.Printf("%-8s %s\n", "0", "(none)")
	} else {
		for i := 0; i < consoleCycleN; i++ {
			cycle := topCycles[i]
			fmt.Printf("%-8d %s\n", len(cycle.Headers), strings.Join(cycle.Headers, " ; "))
		}
	}

	fmt.Printf("Report: %s\n", reportPath)

	return nil
}

func analyzeTranslationUnits(ctx context.Context, workItems []tuWorkItem) ([]tuResult, error) {
	if len(workItems) == 0 {
		return nil, nil
	}

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(workItems) {
		workerCount = len(workItems)
	}

	jobs := make(chan tuWorkItem)
	results := make(chan tuResult, len(workItems))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- analyzeOneTranslationUnit(ctx, job)
			}
		}()
	}

	go func() {
		defer close(results)
		wg.Wait()
	}()

	for _, item := range workItems {
		jobs <- item
	}
	close(jobs)

	out := make([]tuResult, 0, len(workItems))
	for res := range results {
		if res.Err != nil {
			return nil, res.Err
		}
		out = append(out, res)
	}
	return out, nil
}

func analyzeOneTranslationUnit(ctx context.Context, item tuWorkItem) tuResult {
	resolved, tuPath, err := resolveCompileCommand(item.Entry)
	if err != nil {
		return tuResult{
			Err: fmt.Errorf("failed parsing compile command for %s from %s: %w", item.Entry.File, item.DBPath, err),
		}
	}

	deps, err := preprocessDependencies(ctx, resolved)
	if err != nil {
		return tuResult{
			Err: fmt.Errorf("failed preprocessing %s from %s: %w", item.Entry.File, item.DBPath, err),
		}
	}

	return tuResult{
		DBPath:     item.DBPath,
		TUPath:     tuPath,
		Resolved:   resolved,
		Transitive: deps,
	}
}

func fillClosureLinesParallel(ctx context.Context, report []headerImpactEntry, sourceRoots []string, lineCache map[string]int, lineCacheMu *sync.Mutex) (map[string][]string, error) {
	type closureResult struct {
		index        int
		closureLines int
		deps         []string
		err          error
	}

	workIdx := make([]int, 0)
	depGraph := make(map[string][]string)
	for i := range report {
		if report[i].TUCount > 0 && report[i].SampleCompileCmd != nil {
			workIdx = append(workIdx, i)
		}
	}
	if len(workIdx) == 0 {
		return depGraph, nil
	}

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(workIdx) {
		workerCount = len(workIdx)
	}

	jobs := make(chan int)
	results := make(chan closureResult, len(workIdx))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				lines, deps, err := computeHeaderClosureDetails(ctx, report[idx].Path, report[idx].SampleCompileCmd, sourceRoots, lineCache, lineCacheMu)
				results <- closureResult{index: idx, closureLines: lines, deps: deps, err: err}
			}
		}()
	}

	go func() {
		defer close(results)
		wg.Wait()
	}()

	for _, idx := range workIdx {
		jobs <- idx
	}
	close(jobs)

	for res := range results {
		headerPath := report[res.index].Path
		if res.err == nil && res.closureLines > 0 {
			report[res.index].ClosureLines = res.closureLines
		}
		report[res.index].TotalImpact = report[res.index].ClosureLines * report[res.index].TUCount
		depGraph[headerPath] = res.deps
	}
	return depGraph, nil
}

func ensureHeaderEntry(stats map[string]*headerImpactEntry, path string) *headerImpactEntry {
	item := stats[path]
	if item == nil {
		item = &headerImpactEntry{
			Path: path,
			TUs:  make(map[string]struct{}),
		}
		stats[path] = item
	}
	return item
}

func cloneResolvedCompileCommand(in *compileCommandResolved) *compileCommandResolved {
	if in == nil {
		return nil
	}
	out := &compileCommandResolved{
		Compiler:     in.Compiler,
		Directory:    in.Directory,
		SourceArgIdx: in.SourceArgIdx,
	}
	out.Args = append([]string{}, in.Args...)
	return out
}

func mapKeysSorted(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortHeaderImpact(report []headerImpactEntry) {
	sort.Slice(report, func(i, j int) bool {
		if report[i].Impact != report[j].Impact {
			return report[i].Impact > report[j].Impact
		}
		if report[i].TUCount != report[j].TUCount {
			return report[i].TUCount > report[j].TUCount
		}
		if report[i].Lines != report[j].Lines {
			return report[i].Lines > report[j].Lines
		}
		return report[i].Path < report[j].Path
	})
}

func writeHeaderImpactReport(workspacePath string, topTotal []headerImpactEntry, topDirect []headerImpactEntry, topClosure []headerImpactEntry, allHeaders []headerImpactEntry, topTUs []tuImpactEntry, topCycles []headerCycleEntry) (string, error) {
	reportDir := filepath.Join(workspacePath, "reports")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	ts := time.Now().Format("20060102-150405")
	reportPath := filepath.Join(reportDir, fmt.Sprintf("inspect-headers-%s.md", ts))

	f, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create report file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "# Header Impact Report\n\nGenerated: %s\n\n", time.Now().Format(time.RFC3339)); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(f, "## Top 50 Headers By total_impact\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| total_impact | direct_impact | included_tu_count | closure_lines | lines | header |\n|---:|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, item := range topTotal {
		if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %d | %s |\n", item.TotalImpact, item.DirectImpact, item.TUCount, item.ClosureLines, item.Lines, item.Path); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "\n## Top 50 Headers By direct_impact\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| direct_impact | total_impact | included_tu_count | lines | closure_lines | header |\n|---:|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, item := range topDirect {
		if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %d | %s |\n", item.DirectImpact, item.TotalImpact, item.TUCount, item.Lines, item.ClosureLines, item.Path); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "\n## Top 50 Headers By closure_lines\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| closure_lines | total_impact | included_tu_count | lines | header |\n|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, item := range topClosure {
		if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %s |\n", item.ClosureLines, item.TotalImpact, item.TUCount, item.Lines, item.Path); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "\n## Top 50 TUs By Total Preprocessed Lines\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| preprocessed_lines | included_files | translation_unit |\n|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, item := range topTUs {
		if _, err := fmt.Fprintf(f, "| %d | %d | %s |\n", item.PreprocessedLines, item.IncludedFiles, item.TUPath); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "\n## Top 50 Header Dependency Cycles\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| cycle_size | headers |\n|---:|---|\n"); err != nil {
		return "", err
	}
	if len(topCycles) == 0 {
		if _, err := fmt.Fprintf(f, "| 0 | (none) |\n"); err != nil {
			return "", err
		}
	} else {
		for _, cycle := range topCycles {
			if _, err := fmt.Fprintf(f, "| %d | %s |\n", len(cycle.Headers), strings.Join(cycle.Headers, " ; ")); err != nil {
				return "", err
			}
		}
	}

	if _, err := fmt.Fprintf(f, "\n## Translation Units For Top 50 direct_impact Headers\n\n"); err != nil {
		return "", err
	}
	for _, item := range topDirect {
		if _, err := fmt.Fprintf(f, "### %s\n\n", item.Path); err != nil {
			return "", err
		}
		if _, err := fmt.Fprintf(f, "- direct_impact: %d\n- total_impact: %d\n- included_tu_count: %d\n- lines: %d\n- closure_lines: %d\n", item.DirectImpact, item.TotalImpact, item.TUCount, item.Lines, item.ClosureLines); err != nil {
			return "", err
		}
		if len(item.SortedTUs) == 0 {
			if _, err := fmt.Fprintf(f, "- Included By: (none)\n\n"); err != nil {
				return "", err
			}
			continue
		}
		if _, err := fmt.Fprintf(f, "- Included By:\n"); err != nil {
			return "", err
		}
		for _, tu := range item.SortedTUs {
			if _, err := fmt.Fprintf(f, "  - %s\n", tu); err != nil {
				return "", err
			}
		}
		if _, err := fmt.Fprintln(f); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "## All Project Headers And Impact\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| direct_impact | total_impact | included_tu_count | lines | closure_lines | header |\n|---:|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, item := range allHeaders {
		if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %d | %s |\n", item.DirectImpact, item.TotalImpact, item.TUCount, item.Lines, item.ClosureLines, item.Path); err != nil {
			return "", err
		}
	}

	abs, err := filepath.Abs(reportPath)
	if err != nil {
		return reportPath, nil
	}
	return abs, nil
}

func computeTUImpacts(tuResults []tuResult, sourceRoots []string, lineCache map[string]int, lineCacheMu *sync.Mutex) ([]tuImpactEntry, error) {
	impacts := make([]tuImpactEntry, 0, len(tuResults))
	for _, res := range tuResults {
		seen := make(map[string]struct{})
		total := 0
		for _, dep := range res.Transitive {
			depAbs, err := absolutePath(res.Resolved.Directory, dep)
			if err != nil {
				continue
			}
			if !pathInAnyRoot(depAbs, sourceRoots) {
				continue
			}
			if _, ok := seen[depAbs]; ok {
				continue
			}
			seen[depAbs] = struct{}{}
			lines, err := countLinesCached(lineCache, lineCacheMu, depAbs)
			if err != nil {
				continue
			}
			total += lines
		}
		impacts = append(impacts, tuImpactEntry{
			TUPath:            res.TUPath,
			PreprocessedLines: total,
			IncludedFiles:     len(seen),
		})
	}

	sort.Slice(impacts, func(i, j int) bool {
		if impacts[i].PreprocessedLines != impacts[j].PreprocessedLines {
			return impacts[i].PreprocessedLines > impacts[j].PreprocessedLines
		}
		if impacts[i].IncludedFiles != impacts[j].IncludedFiles {
			return impacts[i].IncludedFiles > impacts[j].IncludedFiles
		}
		return impacts[i].TUPath < impacts[j].TUPath
	})

	if len(impacts) > 50 {
		impacts = impacts[:50]
	}
	return impacts, nil
}

func detectHeaderCycles(headers []headerImpactEntry, depGraph map[string][]string) []headerCycleEntry {
	nodeSet := make(map[string]struct{}, len(headers))
	nodes := make([]string, 0, len(headers))
	for _, h := range headers {
		nodeSet[h.Path] = struct{}{}
		nodes = append(nodes, h.Path)
	}
	sort.Strings(nodes)

	adj := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		seen := make(map[string]struct{})
		for _, dep := range depGraph[n] {
			if _, ok := nodeSet[dep]; !ok {
				continue
			}
			if dep == n {
				continue
			}
			if _, ok := seen[dep]; ok {
				continue
			}
			seen[dep] = struct{}{}
			adj[n] = append(adj[n], dep)
		}
		sort.Strings(adj[n])
	}

	index := 0
	indices := make(map[string]int)
	lowlink := make(map[string]int)
	onStack := make(map[string]bool)
	stack := make([]string, 0, len(nodes))
	var components [][]string

	var strongConnect func(v string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adj[v] {
			if _, ok := indices[w]; !ok {
				strongConnect(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] && indices[w] < lowlink[v] {
				lowlink[v] = indices[w]
			}
		}

		if lowlink[v] == indices[v] {
			component := make([]string, 0)
			for {
				last := len(stack) - 1
				w := stack[last]
				stack = stack[:last]
				onStack[w] = false
				component = append(component, w)
				if w == v {
					break
				}
			}
			sort.Strings(component)
			components = append(components, component)
		}
	}

	for _, n := range nodes {
		if _, ok := indices[n]; !ok {
			strongConnect(n)
		}
	}

	cycles := make([]headerCycleEntry, 0)
	for _, comp := range components {
		if len(comp) > 1 {
			cycles = append(cycles, headerCycleEntry{Headers: comp})
		}
	}

	sort.Slice(cycles, func(i, j int) bool {
		if len(cycles[i].Headers) != len(cycles[j].Headers) {
			return len(cycles[i].Headers) > len(cycles[j].Headers)
		}
		return strings.Join(cycles[i].Headers, "\x00") < strings.Join(cycles[j].Headers, "\x00")
	})

	return cycles
}

func limitHeaderCycles(cycles []headerCycleEntry, n int) []headerCycleEntry {
	if n <= 0 || len(cycles) == 0 {
		return []headerCycleEntry{}
	}
	if len(cycles) <= n {
		return cycles
	}
	return cycles[:n]
}

func findAllProjectHeaders(sourceRoots []string) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string

	for _, root := range sourceRoots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !isHeaderFile(path) {
				return nil
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			abs = filepath.Clean(abs)
			if _, ok := seen[abs]; ok {
				return nil
			}
			seen[abs] = struct{}{}
			out = append(out, abs)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed walking source root %s: %w", root, err)
		}
	}

	sort.Strings(out)
	return out, nil
}

func isHeaderFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".h", ".hh", ".hpp", ".hxx", ".h++", ".ipp", ".inl", ".tpp":
		return true
	default:
		return false
	}
}

func inspectToolchains(ctx context.Context, ws *ccommon.WorkspaceContext) ([]string, error) {
	toolchainFlag := ""
	if toolchainFlagRaw := cli.GetOptionalString(ctx, ccommon.ToolchainParameter); toolchainFlagRaw != nil {
		toolchainFlag = *toolchainFlagRaw
	}
	if strings.TrimSpace(toolchainFlag) == "" {
		return ws.ListToolchains(ctx)
	}
	parts := strings.Split(toolchainFlag, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result, nil
}

func inspectConfigs(ctx context.Context, ws *ccommon.WorkspaceContext) []string {
	buildConfigRaw := cli.GetOptionalStringList(ctx, ccommon.ConfigParameter)
	if buildConfigRaw == nil || len(*buildConfigRaw) == 0 {
		return ws.Config.Configurations
	}
	configs := make([]string, 0, len(*buildConfigRaw))
	for _, cfg := range *buildConfigRaw {
		cfg = strings.TrimSpace(cfg)
		if cfg != "" {
			configs = append(configs, cfg)
		}
	}
	return configs
}

func inspectTargets(ctx context.Context, ws *ccommon.WorkspaceContext) ([]string, bool) {
	targetFlagRaw := cli.GetOptionalString(ctx, ccommon.TargetParameter)
	if targetFlagRaw == nil || strings.TrimSpace(*targetFlagRaw) == "" {
		return ws.ListTargets(ctx), false
	}
	parts := strings.Split(*targetFlagRaw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result, true
}

func resolveInspectScope(ctx context.Context, ws *ccommon.WorkspaceContext, selectedTargets []string, hasTargetFilter bool) ([]string, []string, error) {
	if !hasTargetFilter {
		targets := ws.ListTargets(ctx)
		roots, err := inspectSourceRoots(ws)
		if err != nil {
			return nil, nil, err
		}
		return targets, roots, nil
	}

	if len(selectedTargets) == 0 {
		return nil, nil, fmt.Errorf("no targets selected")
	}

	headerRootSet := make(map[string]struct{})
	for _, t := range selectedTargets {
		target, err := ws.GetTarget(ctx, t)
		if err != nil {
			return nil, nil, fmt.Errorf("unknown target %q", t)
		}
		sourcePath, err := target.CMakeSourcePath(ctx, ws)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve source path for target %q: %w", t, err)
		}
		sourcePath, err = filepath.Abs(sourcePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve absolute source path for target %q: %w", t, err)
		}
		headerRootSet[filepath.Clean(sourcePath)] = struct{}{}
	}

	reverseDeps := make(map[string][]string)
	for targetName, cfg := range ws.Config.Targets {
		for _, dep := range cfg.Depends {
			parts := strings.SplitN(dep, "/", 2)
			depTarget := strings.TrimSpace(parts[0])
			if depTarget == "" {
				continue
			}
			reverseDeps[depTarget] = append(reverseDeps[depTarget], targetName)
		}
	}

	visited := make(map[string]struct{})
	queue := make([]string, 0, len(selectedTargets))
	for _, t := range selectedTargets {
		if _, ok := visited[t]; ok {
			continue
		}
		visited[t] = struct{}{}
		queue = append(queue, t)
	}

	for i := 0; i < len(queue); i++ {
		cur := queue[i]
		for _, depd := range reverseDeps[cur] {
			if _, ok := visited[depd]; ok {
				continue
			}
			visited[depd] = struct{}{}
			queue = append(queue, depd)
		}
	}

	tuTargets := make([]string, 0, len(visited))
	for k := range visited {
		tuTargets = append(tuTargets, k)
	}
	sort.Strings(tuTargets)

	headerRoots := make([]string, 0, len(headerRootSet))
	for k := range headerRootSet {
		headerRoots = append(headerRoots, k)
	}
	sort.Strings(headerRoots)

	return tuTargets, headerRoots, nil
}

func inspectSourceRoots(ws *ccommon.WorkspaceContext) ([]string, error) {
	roots := make([]string, 0, len(ws.Config.Sources))
	for sourceName := range ws.Config.Sources {
		sourcePath, err := ws.GetSourcePath(sourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve source path for %s: %w", sourceName, err)
		}
		sourcePath, err = filepath.Abs(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute source path for %s: %w", sourceName, err)
		}
		roots = append(roots, filepath.Clean(sourcePath))
	}
	return roots, nil
}

func inspectCompileCommandsPaths(ctx context.Context, ws *ccommon.WorkspaceContext, toolchains []string, configs []string, targets []string) ([]string, []string, error) {
	seen := make(map[string]struct{})
	var found []string
	var missing []string

	for _, targetName := range targets {
		target, err := ws.GetTarget(ctx, targetName)
		if err != nil {
			return nil, nil, fmt.Errorf("unknown target %q", targetName)
		}
		for _, toolchain := range toolchains {
			for _, config := range configs {
				buildPath, err := target.CMakeBuildPath(ctx, ws, ccommon.TargetBuildParameters{Toolchain: toolchain, BuildType: config})
				if err != nil {
					missing = append(missing, fmt.Sprintf("%s/%s/%s", targetName, toolchain, config))
					continue
				}
				dbPath := filepath.Join(buildPath, "compile_commands.json")
				if _, err := os.Stat(dbPath); err != nil {
					missing = append(missing, dbPath)
					continue
				}
				absPath, err := filepath.Abs(dbPath)
				if err != nil {
					missing = append(missing, dbPath)
					continue
				}
				if _, ok := seen[absPath]; !ok {
					seen[absPath] = struct{}{}
					found = append(found, absPath)
				}
			}
		}
	}

	sort.Strings(found)
	return found, missing, nil
}

func readCompileCommands(path string) ([]compileCommandEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []compileCommandEntry
	if err := json.Unmarshal(content, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func resolveCompileCommand(entry compileCommandEntry) (*compileCommandResolved, string, error) {
	args := append([]string{}, entry.Arguments...)
	if len(args) == 0 {
		var err error
		args, err = splitShellCommand(entry.Command)
		if err != nil {
			return nil, "", err
		}
	}
	if len(args) == 0 {
		return nil, "", fmt.Errorf("empty command line")
	}

	compiler := args[0]
	originalArgs := args[1:]
	sourceArgIdx := findSourceArgIndex(originalArgs, entry.File, entry.Directory)
	tuPath := entry.File
	if tuPath == "" && sourceArgIdx >= 0 {
		tuPath = originalArgs[sourceArgIdx]
	}
	var err error
	tuPath, err = absolutePath(entry.Directory, tuPath)
	if err != nil {
		return nil, "", err
	}

	return &compileCommandResolved{
		Compiler:     compiler,
		Args:         originalArgs,
		Directory:    entry.Directory,
		SourceArgIdx: sourceArgIdx,
	}, tuPath, nil
}

func preprocessDependencies(ctx context.Context, resolved *compileCommandResolved) ([]string, error) {
	depArgs := buildDependencyArgs(resolved.Args)

	cmd := exec.CommandContext(ctx, resolved.Compiler, depArgs...)
	if resolved.Directory != "" {
		cmd.Dir = resolved.Directory
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s %v: %w\n%s", resolved.Compiler, depArgs, err, string(output))
	}

	deps, err := parseMakeDependencies(output)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func computeHeaderClosureDetails(ctx context.Context, headerPath string, resolved *compileCommandResolved, sourceRoots []string, lineCache map[string]int, lineCacheMu *sync.Mutex) (int, []string, error) {
	if resolved == nil || resolved.SourceArgIdx < 0 || resolved.SourceArgIdx >= len(resolved.Args) {
		lines, err := countLinesCached(lineCache, lineCacheMu, headerPath)
		return lines, nil, err
	}
	modifiedArgs := append([]string{}, resolved.Args...)
	modifiedArgs[resolved.SourceArgIdx] = headerPath
	depArgs := buildDependencyArgs(modifiedArgs)

	cmd := exec.CommandContext(ctx, resolved.Compiler, depArgs...)
	if resolved.Directory != "" {
		cmd.Dir = resolved.Directory
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		lines, err2 := countLinesCached(lineCache, lineCacheMu, headerPath)
		return lines, nil, err2
	}
	deps, err := parseMakeDependencies(output)
	if err != nil {
		lines, err2 := countLinesCached(lineCache, lineCacheMu, headerPath)
		return lines, nil, err2
	}

	unique := make(map[string]struct{})
	total := 0
	depList := make([]string, 0)
	for _, dep := range deps {
		depAbs, err := absolutePath(resolved.Directory, dep)
		if err != nil {
			continue
		}
		if !pathInAnyRoot(depAbs, sourceRoots) {
			continue
		}
		if _, ok := unique[depAbs]; ok {
			continue
		}
		unique[depAbs] = struct{}{}
		depList = append(depList, depAbs)
		lines, err := countLinesCached(lineCache, lineCacheMu, depAbs)
		if err != nil {
			continue
		}
		total += lines
	}
	sort.Strings(depList)
	if total == 0 {
		lines, err := countLinesCached(lineCache, lineCacheMu, headerPath)
		return lines, depList, err
	}
	return total, depList, nil
}

func countLinesCached(cache map[string]int, mu *sync.Mutex, path string) (int, error) {
	mu.Lock()
	if v, ok := cache[path]; ok {
		mu.Unlock()
		return v, nil
	}
	mu.Unlock()

	v, err := countLines(path)
	if err != nil {
		return 0, err
	}

	mu.Lock()
	cache[path] = v
	mu.Unlock()
	return v, nil
}

func buildDependencyArgs(args []string) []string {
	filtered := make([]string, 0, len(args)+4)
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}

		switch arg {
		case "-c", "-E", "-M", "-MM", "-MD", "-MMD", "-MG", "-MP":
			continue
		case "-o", "-MF", "-MT", "-MQ", "-MJ":
			skipNext = true
			continue
		}

		if strings.HasPrefix(arg, "-MF") || strings.HasPrefix(arg, "-MT") || strings.HasPrefix(arg, "-MQ") || strings.HasPrefix(arg, "-MJ") {
			continue
		}
		if strings.HasPrefix(arg, "-o") && len(arg) > 2 {
			continue
		}

		filtered = append(filtered, arg)
	}

	filtered = append(filtered, "-w", "-MM")
	return filtered
}

func findSourceArgIndex(args []string, entryFile string, dir string) int {
	if entryFile != "" {
		entryAbs, err := absolutePath(dir, entryFile)
		if err == nil {
			for i, arg := range args {
				if strings.HasPrefix(arg, "-") {
					continue
				}
				argAbs, err := absolutePath(dir, arg)
				if err == nil && argAbs == entryAbs {
					return i
				}
			}
		}
	}
	for i := len(args) - 1; i >= 0; i-- {
		arg := strings.TrimSpace(args[i])
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		return i
	}
	return -1
}

func parseMakeDependencies(output []byte) ([]string, error) {
	normalized := strings.ReplaceAll(string(output), "\\\n", " ")
	normalized = strings.ReplaceAll(normalized, "\r", "")
	lines := strings.Split(normalized, "\n")

	ruleLine := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, ":") {
			ruleLine = line
			break
		}
	}
	if ruleLine == "" {
		return nil, fmt.Errorf("dependency output did not contain a make-style dependency rule")
	}

	colon := strings.Index(ruleLine, ":")
	if colon < 0 || colon == len(ruleLine)-1 {
		return nil, fmt.Errorf("dependency output was malformed")
	}

	depsPart := strings.TrimSpace(ruleLine[colon+1:])
	deps := splitEscapedFields(depsPart)
	if len(deps) == 0 {
		return nil, fmt.Errorf("dependency output had no dependencies")
	}

	return deps, nil
}

func splitEscapedFields(input string) []string {
	fields := []string{}
	var current strings.Builder
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == ' ' || r == '\t' || r == '\n' {
			flush()
			continue
		}
		current.WriteRune(r)
	}
	if escaped {
		current.WriteRune('\\')
	}
	flush()
	return fields
}

func splitShellCommand(command string) ([]string, error) {
	args := []string{}
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for i := 0; i < len(command); i++ {
		ch := command[i]
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			if inSingle {
				current.WriteByte(ch)
			} else {
				escaped = true
			}
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if (ch == ' ' || ch == '\t' || ch == '\n') && !inSingle && !inDouble {
			flush()
			continue
		}

		current.WriteByte(ch)
	}

	if escaped {
		current.WriteByte('\\')
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote in command: %s", command)
	}
	flush()
	return args, nil
}

func findSourcePathFromArgs(args []string) string {
	for i := len(args) - 1; i >= 0; i-- {
		arg := strings.TrimSpace(args[i])
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		return arg
	}
	return ""
}

func absolutePath(baseDir string, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("missing path")
	}
	p := path
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}
	p, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(p), nil
}

func pathInAnyRoot(path string, roots []string) bool {
	for _, root := range roots {
		if pathWithinRoot(path, root) {
			return true
		}
	}
	return false
}

func pathWithinRoot(path string, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 16*1024*1024)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
