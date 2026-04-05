package cbuildapp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

type symbolStats struct {
	Name          string
	TotalCount    int
	TUCount       int
	Score         int
	CompiledBytes uint64
	ImpactBytes   uint64
	Objects       map[string]struct{}
	ObjectBytes   map[string]uint64
	SortedObjects []string
}

type objectSymbolResult struct {
	ObjectPath string
	Symbols    []parsedSymbol
	HasSize    bool
	Err        error
}

type parsedSymbol struct {
	Name       string
	Size       uint64
	Address    uint64
	HasAddress bool
	Type       byte
}

type nmMode struct {
	Binary string
	Args   []string
}

func runInspectSymbols(ctx context.Context, args []string) error {
	workspacePath := "."
	if workspacePathRaw := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter); workspacePathRaw != nil && *workspacePathRaw != "" {
		workspacePath = *workspacePathRaw
	}

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchains, err := inspectToolchains(ctx, ws)
	if err != nil {
		return err
	}
	if len(toolchains) == 0 {
		return fmt.Errorf("no toolchains found")
	}

	configs := inspectConfigs(ctx, ws)
	targets, _ := inspectTargets(ctx, ws)

	objectFiles, missingBuildDirs, err := collectObjectFiles(ctx, ws, toolchains, configs, targets)
	if err != nil {
		return err
	}
	if len(objectFiles) == 0 {
		return fmt.Errorf("no object files found for selected targets/toolchains/configs; run cbuild build first")
	}
	if len(missingBuildDirs) > 0 {
		fmt.Printf("Skipping %d build directories not found\n", len(missingBuildDirs))
	}

	symbolMap, hasSizes, err := analyzeObjectSymbols(ctx, objectFiles)
	if err != nil {
		return err
	}
	if !hasSizes {
		fmt.Println("Warning: symbol sizes were not available from nm; impact_bytes may be zero.")
	}

	allRepeated := make([]symbolStats, 0)
	for _, stat := range symbolMap {
		stat.TUCount = len(stat.Objects)
		if stat.TUCount <= 1 {
			continue
		}
		stat.Score = stat.TUCount * stat.TotalCount
		stat.ImpactBytes = stat.CompiledBytes * uint64(stat.TUCount)
		stat.SortedObjects = mapKeysSorted(stat.Objects)
		allRepeated = append(allRepeated, *stat)
	}

	if len(allRepeated) == 0 {
		fmt.Println("No repeated symbols found across multiple translation units.")
		return nil
	}

	sort.Slice(allRepeated, func(i, j int) bool {
		if allRepeated[i].ImpactBytes != allRepeated[j].ImpactBytes {
			return allRepeated[i].ImpactBytes > allRepeated[j].ImpactBytes
		}
		if allRepeated[i].CompiledBytes != allRepeated[j].CompiledBytes {
			return allRepeated[i].CompiledBytes > allRepeated[j].CompiledBytes
		}
		if allRepeated[i].Score != allRepeated[j].Score {
			return allRepeated[i].Score > allRepeated[j].Score
		}
		if allRepeated[i].TUCount != allRepeated[j].TUCount {
			return allRepeated[i].TUCount > allRepeated[j].TUCount
		}
		if allRepeated[i].TotalCount != allRepeated[j].TotalCount {
			return allRepeated[i].TotalCount > allRepeated[j].TotalCount
		}
		return allRepeated[i].Name < allRepeated[j].Name
	})

	nonStdRepeated := make([]symbolStats, 0, len(allRepeated))
	for _, item := range allRepeated {
		if isStdLikeSymbol(item.Name) {
			continue
		}
		nonStdRepeated = append(nonStdRepeated, item)
	}

	topN := 50
	if len(allRepeated) < topN {
		topN = len(allRepeated)
	}
	topRepeated := allRepeated[:topN]
	topNonStdN := 50
	if len(nonStdRepeated) < topNonStdN {
		topNonStdN = len(nonStdRepeated)
	}
	topNonStd := nonStdRepeated[:topNonStdN]

	reportPath, err := writeInspectSymbolsReport(workspacePath, topNonStd, topRepeated, allRepeated)
	if err != nil {
		return err
	}

	consoleN := 20
	if len(topNonStd) < consoleN {
		consoleN = len(topNonStd)
	}
	fmt.Printf("Top %d non-std repeated symbols across translation units\n", consoleN)
	fmt.Printf("%-14s %-14s %-10s %-18s %s\n", "impact_bytes", "compiled_bytes", "tu_count", "total_occurrences", "symbol")
	for _, item := range topNonStd[:consoleN] {
		fmt.Printf("%-14d %-14d %-10d %-18d %s\n", item.ImpactBytes, item.CompiledBytes, item.TUCount, item.TotalCount, item.Name)
	}

	consoleAllN := 20
	if len(topRepeated) < consoleAllN {
		consoleAllN = len(topRepeated)
	}
	fmt.Printf("\nTop %d repeated symbols across translation units\n", consoleAllN)
	fmt.Printf("%-14s %-14s %-10s %-18s %s\n", "impact_bytes", "compiled_bytes", "tu_count", "total_occurrences", "symbol")
	for _, item := range topRepeated[:consoleAllN] {
		fmt.Printf("%-14d %-14d %-10d %-18d %s\n", item.ImpactBytes, item.CompiledBytes, item.TUCount, item.TotalCount, item.Name)
	}
	fmt.Printf("Report: %s\n", reportPath)

	return nil
}

func collectObjectFiles(ctx context.Context, ws *ccommon.WorkspaceContext, toolchains []string, configs []string, targets []string) ([]string, []string, error) {
	seen := make(map[string]struct{})
	objectFiles := make([]string, 0)
	missingBuildDirs := make([]string, 0)

	for _, targetName := range targets {
		target, err := ws.GetTarget(ctx, targetName)
		if err != nil {
			return nil, nil, fmt.Errorf("unknown target %q", targetName)
		}

		for _, toolchain := range toolchains {
			for _, config := range configs {
				buildPath, err := target.CMakeBuildPath(ctx, ws, ccommon.TargetBuildParameters{
					Toolchain: toolchain,
					BuildType: config,
				})
				if err != nil {
					return nil, nil, fmt.Errorf("failed to resolve build path for %s/%s/%s: %w", targetName, toolchain, config, err)
				}

				if _, err := os.Stat(buildPath); err != nil {
					missingBuildDirs = append(missingBuildDirs, buildPath)
					continue
				}

				err = filepath.WalkDir(buildPath, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					ext := strings.ToLower(filepath.Ext(path))
					if ext != ".o" && ext != ".obj" {
						return nil
					}
					absPath, err := filepath.Abs(path)
					if err != nil {
						return err
					}
					absPath = filepath.Clean(absPath)
					if _, ok := seen[absPath]; ok {
						return nil
					}
					seen[absPath] = struct{}{}
					objectFiles = append(objectFiles, absPath)
					return nil
				})
				if err != nil {
					return nil, nil, fmt.Errorf("failed walking build directory %s: %w", buildPath, err)
				}
			}
		}
	}

	sort.Strings(objectFiles)
	sort.Strings(missingBuildDirs)
	return objectFiles, missingBuildDirs, nil
}

func analyzeObjectSymbols(ctx context.Context, objectFiles []string) (map[string]*symbolStats, bool, error) {
	symbolMap := make(map[string]*symbolStats)
	if len(objectFiles) == 0 {
		return symbolMap, false, nil
	}

	modes := nmCandidateModes()

	firstObj := objectFiles[0]
	primaryMode, firstOutput, firstHasSize, err := detectNMMode(ctx, firstObj, modes)
	if err != nil {
		return nil, false, err
	}
	firstSymbols := parseNMDefinedSymbols(firstOutput)
	addObjectSymbols(symbolMap, firstObj, firstSymbols)
	hasAnySize := firstHasSize || hasAnySizedSymbols(firstSymbols)

	if len(objectFiles) == 1 {
		return symbolMap, hasAnySize, nil
	}

	remaining := objectFiles[1:]
	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(remaining) {
		workerCount = len(remaining)
	}

	jobs := make(chan string)
	results := make(chan objectSymbolResult, len(remaining))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for obj := range jobs {
				out, err := runNMWithFallback(ctx, obj, primaryMode, modes)
				if err != nil {
					results <- objectSymbolResult{ObjectPath: obj, Err: err}
					continue
				}
				syms := parseNMDefinedSymbols(out)
				results <- objectSymbolResult{ObjectPath: obj, Symbols: syms, HasSize: hasAnySizedSymbols(syms)}
			}
		}()
	}

	go func() {
		defer close(results)
		wg.Wait()
	}()

	for _, obj := range remaining {
		jobs <- obj
	}
	close(jobs)

	for res := range results {
		if res.Err != nil {
			return nil, false, res.Err
		}
		if res.HasSize {
			hasAnySize = true
		}
		addObjectSymbols(symbolMap, res.ObjectPath, res.Symbols)
	}

	return symbolMap, hasAnySize, nil
}

func nmCandidateModes() []nmMode {
	// Prefer llvm-nm first for better cross-platform size support; then system nm.
	binaries := []string{"llvm-nm", "nm"}
	argSets := [][]string{
		{"-C", "-S", "--defined-only"},
		{"-C", "--print-size", "--defined-only"},
		{"-C", "-S", "-U"},
		{"-C", "--defined-only"},
		{"-C", "-U"},
		{"-C"},
		{},
	}
	out := make([]nmMode, 0, len(binaries)*len(argSets))
	for _, bin := range binaries {
		for _, args := range argSets {
			out = append(out, nmMode{Binary: bin, Args: args})
		}
	}
	return out
}

func detectNMMode(ctx context.Context, objectFile string, candidates []nmMode) (nmMode, []byte, bool, error) {
	var fallbackMode nmMode
	var fallbackOutput []byte
	hasFallback := false
	var lastErr error
	for _, c := range candidates {
		out, err := runNM(ctx, c.Binary, objectFile, c.Args)
		if err == nil {
			syms := parseNMDefinedSymbols(out)
			if hasAnySizedSymbols(syms) {
				return c, out, true, nil
			}
			if !hasFallback {
				fallbackMode = c
				fallbackOutput = out
				hasFallback = true
			}
			continue
		}
		lastErr = err
	}
	if hasFallback {
		return fallbackMode, fallbackOutput, false, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("failed to execute nm")
	}
	return nmMode{}, nil, false, lastErr
}

func runNMWithFallback(ctx context.Context, objectFile string, primary nmMode, candidates []nmMode) ([]byte, error) {
	out, err := runNM(ctx, primary.Binary, objectFile, primary.Args)
	if err == nil {
		return out, nil
	}

	for _, c := range candidates {
		if sameNMMode(c, primary) {
			continue
		}
		out, err2 := runNM(ctx, c.Binary, objectFile, c.Args)
		if err2 == nil {
			return out, nil
		}
	}
	return nil, err
}

func runNM(ctx context.Context, binary string, objectFile string, args []string) ([]byte, error) {
	cmdArgs := make([]string, 0, len(args)+1)
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, objectFile)

	cmd := exec.CommandContext(ctx, binary, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run %s on %s with args %v: %w\n%s", binary, objectFile, args, err, string(output))
	}
	return output, nil
}

func hasAnySizedSymbols(symbols []parsedSymbol) bool {
	for _, s := range symbols {
		if s.Size > 0 {
			return true
		}
	}
	return false
}

func parseNMDefinedSymbols(output []byte) []parsedSymbol {
	lines := strings.Split(strings.ReplaceAll(string(output), "\r", ""), "\n")
	allSymbols := make([]parsedSymbol, 0)
	includeInReport := make([]bool, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "nm:") || strings.Contains(line, "no symbols") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		start := 0
		if strings.HasSuffix(fields[0], ":") {
			start = 1
			if len(fields)-start < 2 {
				continue
			}
		}

		var typ string
		var name string
		var size uint64
		var address uint64
		var hasAddress bool

		switch rem := len(fields) - start; {
		case rem >= 4:
			// <addr> <size> <type> <name...>
			if parsed, ok := parseUintToken(fields[start]); ok {
				address = parsed
				hasAddress = true
			}
			if parsed, ok := parseUintToken(fields[start+1]); ok {
				size = parsed
				typ = fields[start+2]
				name = strings.Join(fields[start+3:], " ")
			} else {
				typ = fields[start+1]
				name = strings.Join(fields[start+2:], " ")
			}
		case rem >= 3:
			// <addr> <type> <name...>
			if parsed, ok := parseUintToken(fields[start]); ok {
				address = parsed
				hasAddress = true
			}
			typ = fields[start+1]
			name = strings.Join(fields[start+2:], " ")
		default:
			typ = fields[start]
			name = strings.Join(fields[start+1:], " ")
		}

		typ = strings.TrimSpace(typ)
		name = strings.TrimSpace(name)
		if len(typ) != 1 || name == "" {
			continue
		}
		t := typ[len(typ)-1]
		if (t < 'A' || t > 'Z') && (t < 'a' || t > 'z') {
			continue
		}
		if t == 'U' || t == 'u' || t == '?' {
			continue
		}
		entry := parsedSymbol{
			Name:       name,
			Size:       size,
			Address:    address,
			HasAddress: hasAddress,
			Type:       t,
		}
		allSymbols = append(allSymbols, entry)

		// Keep local symbols out of results, but still use them as address boundaries.
		include := t >= 'A' && t <= 'Z' && !isJunkSymbolName(name)
		includeInReport = append(includeInReport, include)
	}
	estimateZeroSizesFromAddresses(allSymbols)

	symbols := make([]parsedSymbol, 0, len(allSymbols))
	for i, sym := range allSymbols {
		if !includeInReport[i] {
			continue
		}
		symbols = append(symbols, sym)
	}
	return symbols
}

func estimateZeroSizesFromAddresses(symbols []parsedSymbol) {
	if len(symbols) < 2 {
		return
	}

	buckets := make(map[string][]int)
	for i := range symbols {
		if !symbols[i].HasAddress {
			continue
		}
		bucket := symbolAddressBucket(symbols[i].Type)
		// Address-delta sizing is reliable enough for code symbols, but can wildly
		// overestimate sparse data symbols (e.g., guard variables), so limit fallback
		// estimation to text/code buckets.
		if bucket != "text" {
			continue
		}
		buckets[bucket] = append(buckets[bucket], i)
	}

	for _, idxs := range buckets {
		if len(idxs) < 2 {
			continue
		}
		sort.Slice(idxs, func(i, j int) bool {
			lhs := symbols[idxs[i]]
			rhs := symbols[idxs[j]]
			if lhs.Address != rhs.Address {
				return lhs.Address < rhs.Address
			}
			return lhs.Name < rhs.Name
		})

		for pos, idx := range idxs {
			if symbols[idx].Size > 0 {
				continue
			}
			cur := symbols[idx].Address
			for j := pos + 1; j < len(idxs); j++ {
				nextAddr := symbols[idxs[j]].Address
				if nextAddr > cur {
					symbols[idx].Size = nextAddr - cur
					break
				}
			}
		}
	}
}

func symbolAddressBucket(typ byte) string {
	upper := typ
	if upper >= 'a' && upper <= 'z' {
		upper = upper - ('a' - 'A')
	}
	switch upper {
	case 'T', 'W', 'I':
		return "text"
	case 'D', 'S', 'B', 'R', 'V', 'G', 'C':
		return "data"
	default:
		return string(typ)
	}
}

func parseUintToken(token string) (uint64, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, false
	}
	if strings.HasPrefix(token, "0x") || strings.HasPrefix(token, "0X") {
		v, err := strconv.ParseUint(token[2:], 16, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}
	// Try hex first, then decimal.
	if v, err := strconv.ParseUint(token, 16, 64); err == nil {
		return v, true
	}
	if v, err := strconv.ParseUint(token, 10, 64); err == nil {
		return v, true
	}
	return 0, false
}

func addObjectSymbols(symbolMap map[string]*symbolStats, objectPath string, symbols []parsedSymbol) {
	seenInObj := make(map[string]uint64)
	for _, sym := range symbols {
		entry := symbolMap[sym.Name]
		if entry == nil {
			entry = &symbolStats{
				Name:        sym.Name,
				Objects:     make(map[string]struct{}),
				ObjectBytes: make(map[string]uint64),
			}
			symbolMap[sym.Name] = entry
		}
		entry.TotalCount++
		if current, ok := seenInObj[sym.Name]; !ok || sym.Size > current {
			seenInObj[sym.Name] = sym.Size
		}
	}

	for symName, size := range seenInObj {
		entry := symbolMap[symName]
		entry.Objects[objectPath] = struct{}{}
		if existing, ok := entry.ObjectBytes[objectPath]; !ok || size > existing {
			entry.ObjectBytes[objectPath] = size
		}
		if size > entry.CompiledBytes {
			entry.CompiledBytes = size
		}
	}
}

func sameStringSlice(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameNMMode(a nmMode, b nmMode) bool {
	return a.Binary == b.Binary && sameStringSlice(a.Args, b.Args)
}

func isJunkSymbolName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return true
	}
	if strings.HasPrefix(name, "ltmp") && isDigits(name[len("ltmp"):]) {
		return true
	}
	if strings.HasPrefix(name, "Ltmp") && isDigits(name[len("Ltmp"):]) {
		return true
	}
	if strings.HasPrefix(name, "l_.") || strings.HasPrefix(name, ".L") {
		return true
	}
	return false
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isStdLikeSymbol(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	// Common mangled prefixes for std::* symbols (Itanium ABI).
	if strings.HasPrefix(name, "_ZSt") || strings.HasPrefix(name, "_ZNSt") || strings.HasPrefix(name, "_ZNKSt") || strings.HasPrefix(name, "_ZTSt") {
		return true
	}

	return firstTopLevelNamespace(name) == "std"
}

func firstTopLevelNamespace(symbol string) string {
	s := strings.TrimSpace(symbol)
	if s == "" {
		return ""
	}

	angleDepth := 0
	parenDepth := 0
	bracketDepth := 0

	for i := 0; i+1 < len(s); i++ {
		switch s[i] {
		case '<':
			angleDepth++
		case '>':
			if angleDepth > 0 {
				angleDepth--
			}
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		}

		if s[i] == ':' && s[i+1] == ':' && angleDepth == 0 && parenDepth == 0 && bracketDepth == 0 {
			j := i - 1
			for j >= 0 && (s[j] == ' ' || s[j] == '\t' || s[j] == '*' || s[j] == '&') {
				j--
			}
			if j < 0 {
				return ""
			}
			end := j + 1
			for j >= 0 && isIdentByte(s[j]) {
				j--
			}
			ns := s[j+1 : end]
			ns = strings.TrimPrefix(ns, "~")
			return ns
		}
	}
	return ""
}

func isIdentByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func writeInspectSymbolsReport(workspacePath string, topNonStd []symbolStats, topSymbols []symbolStats, allSymbols []symbolStats) (string, error) {
	reportDir := filepath.Join(workspacePath, "reports")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	ts := time.Now().Format("20060102-150405")
	reportPath := filepath.Join(reportDir, fmt.Sprintf("inspect-symbols-%s.md", ts))

	f, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create report file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "# Symbol Repeat Report\n\nGenerated: %s\n\n", time.Now().Format(time.RFC3339)); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(f, "## Top 50 Non-std Repeated Symbols Across Translation Units\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| impact_bytes | compiled_bytes | score | tu_count | total_occurrences | symbol |\n|---:|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	if len(topNonStd) == 0 {
		if _, err := fmt.Fprintf(f, "| 0 | 0 | 0 | 0 | 0 | (none) |\n"); err != nil {
			return "", err
		}
	} else {
		for _, sym := range topNonStd {
			if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %d | %s |\n", sym.ImpactBytes, sym.CompiledBytes, sym.Score, sym.TUCount, sym.TotalCount, sym.Name); err != nil {
				return "", err
			}
		}
	}

	if _, err := fmt.Fprintf(f, "## Top 50 Repeated Symbols Across Translation Units\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| impact_bytes | compiled_bytes | score | tu_count | total_occurrences | symbol |\n|---:|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, sym := range topSymbols {
		if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %d | %s |\n", sym.ImpactBytes, sym.CompiledBytes, sym.Score, sym.TUCount, sym.TotalCount, sym.Name); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "\n## Object Files For Top 50 Symbols\n\n"); err != nil {
		return "", err
	}
	for _, sym := range topSymbols {
		if _, err := fmt.Fprintf(f, "### %s\n\n- impact_bytes: %d\n- compiled_bytes: %d\n- score: %d\n- tu_count: %d\n- total_occurrences: %d\n- object_files:\n", sym.Name, sym.ImpactBytes, sym.CompiledBytes, sym.Score, sym.TUCount, sym.TotalCount); err != nil {
			return "", err
		}
		for _, obj := range sym.SortedObjects {
			objSize := sym.ObjectBytes[obj]
			if _, err := fmt.Fprintf(f, "  - %s (%d bytes)\n", obj, objSize); err != nil {
				return "", err
			}
		}
		if _, err := fmt.Fprintln(f); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(f, "## All Repeated Symbols\n\n"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(f, "| impact_bytes | compiled_bytes | score | tu_count | total_occurrences | symbol |\n|---:|---:|---:|---:|---:|---|\n"); err != nil {
		return "", err
	}
	for _, sym := range allSymbols {
		if _, err := fmt.Fprintf(f, "| %d | %d | %d | %d | %d | %s |\n", sym.ImpactBytes, sym.CompiledBytes, sym.Score, sym.TUCount, sym.TotalCount, sym.Name); err != nil {
			return "", err
		}
	}

	absPath, err := filepath.Abs(reportPath)
	if err != nil {
		return reportPath, nil
	}
	return absPath, nil
}
