package pkg

import (
	"fmt"
	"github.com/barasher/go-exiftool"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

const (
	unknown = "unknown"
)

type Storage struct {
	Entries        []os.DirEntry
	Categories     map[string][]string // [categories][]extensions
	Extensions     map[string]string   // [extensions][categories]
	OutDirectories map[string][]string // []categories[files]
	SubDirs        map[string][]string // [subDir][]extensions
	Unprocessed    []string
	SortMap        map[string]string //image:year, videos:month, documents:month
	Exif           *exiftool.Exiftool
}

func NewStorage() *Storage {
	return &Storage{
		Categories:     make(map[string][]string),
		Extensions:     make(map[string]string),
		OutDirectories: make(map[string][]string),
		SubDirs:        make(map[string][]string),
		Unprocessed:    make([]string, 0),
		SortMap:        make(map[string]string),
	}
}

type Operator struct {
	Storage        Storage
	Flags          Flags
	CsvHandler     *CSVLogger
	SubDirCount    int
	ExtensionCount int
	sem            chan struct{}
	once           sync.Once
	mu             sync.Mutex
}

const defaultPoolSize = 8

func (o *Operator) initPool() {
	o.once.Do(func() {
		o.sem = make(chan struct{}, defaultPoolSize)
	})
}

func GetNewOperator() (*Operator, error) {
	o := &Operator{
		Storage:        *NewStorage(),
		Flags:          Flags{},
		CsvHandler:     nil,
		SubDirCount:    0,
		ExtensionCount: 0,
		sem:            nil,
		once:           sync.Once{},
		mu:             sync.Mutex{},
	}
	o.initPool()

	var err error
	o.Storage.Exif, err = initExifTool()
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Operator) BuildStorageMaps(c *Config) {
	for _, rule := range c.Rules {
		o.Storage.Categories[rule.Category] = make([]string, 0)
		for _, extension := range rule.Extensions {
			o.Storage.Categories[rule.Category] = append(o.Storage.Categories[rule.Category], extension)
			o.Storage.Extensions[extension] = rule.Category
		}
		if rule.SeparateExists() {
			o.Storage.SubDirs[rule.Category] = append(o.Storage.SubDirs[rule.Category], rule.Separate...)
		}
		if rule.Sort != "" {
			o.Storage.SortMap[rule.Category] = rule.Sort
		}
	}
}

func (o *Operator) GetSeparateSubdirs(category, ext string) string {
	if subdirs, exists := o.Storage.SubDirs[category]; exists {
		for _, sub := range subdirs {
			if sub == ext {
				return sub
			}
		}

		return ""
	}

	return ""
}

func (o *Operator) GetSortSubDirs(category string) (string, bool) {
	if sortType, exists := o.Storage.SortMap[category]; exists {
		return sortType, true
	}

	return "", false
}

func (o *Operator) GetExtensionCategory(ext string) (string, bool) {
	if val, ok := o.Storage.Extensions[ext]; ok {
		return val, true
	}

	return unknown, false
}

// AddType adds and returns category of the file
func (o *Operator) AddType(ext, fp string) string {
	category, exists := o.GetExtensionCategory(ext)
	if !exists {
		slog.Warn("unknown extension, doesn't match to rules", "extension", ext)
		slog.Warn("copying to the unknown dir", "filepath", fp)

		return unknown
	}
	o.Storage.OutDirectories[category] = append(o.Storage.OutDirectories[category], fp)

	return category
}

func (o *Operator) CreateSubdirs(dstBasePath string, rules []Rule) error {
	if o.Flags.DryRun {
		return nil
	}

	if err := createDirectory(dstBasePath); err != nil {
		return err
	}

	for _, rule := range rules {
		if err := os.Mkdir(path.Join(dstBasePath, rule.Category), syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY); err != nil {
			return err
		}
		if rule.SeparateExists() {
			for _, separateDir := range rule.Separate {
				if err := os.Mkdir(path.Join(dstBasePath, rule.Category, separateDir), syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// uniqueDstPath has two tasks.
// original task: if there's two file with same name, to not overwriting, add an '_' and number depending on how many copies do exist.
// task that got added during sort-image-files, which will be refactored and improved,
// is to create YEAR, and YEAR/MONTH directories if they don't exist. Q: why is it done here currently?
// because getFileDate function returns the format as in YEAR-MONTH and
func uniqueDstPath(dstBasePath, dstDir, specialDir, baseName string) string {
	ext := filepath.Ext(baseName)
	base := strings.TrimSuffix(baseName, ext)
	dstNewPath := path.Join(dstBasePath, dstDir, baseName)
	if specialDir != "" {
		dstNewPath = path.Join(dstBasePath, dstDir, specialDir, baseName)
		// create the specialDir if it doesn't exist. this is only required for year/month sort things.
		if err := createDirectory(path.Join(dstBasePath, dstDir, specialDir)); err != nil {
			panic(err)
		}
	}

	// TODO: improve this following idiotic logic
	original := dstNewPath
	i := 1
	for {
		if _, err := os.Stat(dstNewPath); err != nil {
			if os.IsNotExist(err) {
				break
			}
			slog.Error("stat call failed during trying to create a unique destination path", "PATH:", dstNewPath)
			panic(err)
		}
		// FIXME: wtf is going on here xD
		// if specialDir == "" {
		// dstNewPath = path.Join(path.Dir(original), fmt.Sprintf("%s_%d%s", base, i, ext))
		// } else {
		dstNewPath = path.Join(path.Dir(original), fmt.Sprintf("%s_%d%s", base, i, ext))
		//}
		i++
	}

	return dstNewPath
}

func (o *Operator) Copy(dstPath, dstDir, specialDir, fileAbsolutePath string) error {
	srcFile, err := os.Open(fileAbsolutePath)
	if err != nil {
		slog.Warn("Skipping unreadable file", "path", fileAbsolutePath, "error", err)
		// o.Storage.Unprocessed = append(o.Storage.Unprocessed, fileAbsolutePath)
		// TODO: why is this commented out? check out later how do we deal with this.
		return nil
	}
	defer func() {
		err := srcFile.Close()
		if err != nil {
			panic(fmt.Errorf("failed to close:%s:%w", fileAbsolutePath, err))
		}
	}()

	_, fileName := path.Split(fileAbsolutePath)
	destinationFile, err := os.Create(uniqueDstPath(dstPath, dstDir, specialDir, fileName))
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			panic(err)
		}
	}(destinationFile)

	_, err = io.Copy(destinationFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy %s file to %s: %w", srcFile.Name(), destinationFile.Name(), err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file:%s:%w", destinationFile.Name(), err)
	}

	if o.CsvHandler != nil {
		if err := o.CsvHandler.Log("SUCCESS", srcFile.Name(), fileName, destinationFile.Name()); err != nil {
			slog.Error("failure-log", "error", err.Error())
		}
	}

	return nil
}

// skipcheck logs skipped files and adds them to unprocessed slice.
func (o *Operator) skipcheck(fp string) bool {
	info, err := os.Stat(fp)
	if err != nil {
		slog.Warn("Skipping blocked file", "path", fp, "error", err)
		o.Storage.Unprocessed = append(o.Storage.Unprocessed, fp)

		return true
	}
	if !info.Mode().IsRegular() {
		slog.Warn("Skipping blocked file", "path", fp, "error", "isn't a regular file")
		o.Storage.Unprocessed = append(o.Storage.Unprocessed, fp)

		return true
	}

	if info.Size() == 0 {
		slog.Warn("Skipping blocked file", "path", fp, "error", "has size 0")
		o.Storage.Unprocessed = append(o.Storage.Unprocessed, fp)

		return true
	}

	return false
}

func (o *Operator) getSpecialSubDirNames(typeDir, ext, fp string) (string, error) {
	// special subDir is what you define in category as part of rules
	specialSubDir := o.GetSeparateSubdirs(typeDir, ext)
	// get the file date depending on sortDir=year/month and pass it to o.Copy
	sortDir, exists := o.GetSortSubDirs(typeDir)
	var err error
	if exists {
		sortDir, err = o.getFileDate(fp, sortDir)
		// if the error is because we couldn't get exif date, then ignore the error
		// otherwise return error.
		if err != nil && !errors.Is(err, ErrNoCreateDate) {
			return "", err
		}
		if sortDir != "" {
			specialSubDir = path.Join(specialSubDir, sortDir)
		}
	}

	return specialSubDir, nil
}

const defaultSemLimit = 10

func (o *Operator) AsyncProcessDir(dirpath string, r bool) (int, error) {
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return 0, err
	}
	slog.Debug("", "entry count:", len(entries))
	if o.Flags.DryRun {
		os.Exit(1)
	}
	total := len(entries)
	processed := int64(0)
	extensions := make([]string, 0)
	var extMutex, unprocMutex = sync.Mutex{}, sync.Mutex{}
	sem := make(chan struct{}, defaultSemLimit)
	var wg sync.WaitGroup

	for _, entry := range entries {
		fp := path.Join(dirpath, entry.Name())
		if entry.IsDir() {
			o.SubDirCount++
			if _, err := o.AsyncProcessDir(fp, true); err != nil {
				return 0, err
			}

			continue
		}
		if o.skipcheck(fp) {
			continue
		}

		kind := path.Ext(fp)
		ext := ""
		if kind != "" {
			ext = kind[1:]
		}

		typeDir := o.AddType(ext, fp)

		wg.Add(1)
		sem <- struct{}{} // get slot
		// TODO how do we handle errors in go calls, can we still just return them?
		go func(fp, typeDir string, ext string) {
			defer wg.Done()
			defer func() { <-sem }() // release slot
			specialSubDir, err := o.getSpecialSubDirNames(typeDir, ext, fp)
			if err != nil {
				return
			}
			if err := o.Copy(o.Flags.DstPath, typeDir, specialSubDir, fp); err != nil {
				unprocMutex.Lock()
				o.Storage.Unprocessed = append(o.Storage.Unprocessed, fp)
				unprocMutex.Unlock()

				return
			}

			atomic.AddInt64(&processed, 1)
			if !r {
				pct := float64(atomic.LoadInt64(&processed)) / float64(total) * 100 //nolint:mnd
				if atomic.LoadInt64(&processed)%int64(max(1, total/20)) == 0 {      //nolint:mnd
					slog.Info("progress", "completed", fmt.Sprintf("%.1f%%", pct))
				}
			}
			extMutex.Lock()
			extensions = append(extensions, ext)
			extMutex.Unlock()
		}(fp, typeDir, ext)
	}
	wg.Wait()
	extensions = RemoveDuplicateStr(extensions)

	return len(extensions), nil
}

func (o *Operator) ProcessDir(dirpath string, r bool) (int, error) {
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return 0, err
	}
	slog.Info("", "entry count:", len(entries))
	if o.Flags.DryRun {
		os.Exit(1)
	}

	total := len(entries)
	processed := 0
	subDirCount := 0
	extensions := make([]string, 0)
	for _, entry := range entries {
		fp := path.Join(dirpath, entry.Name())
		if entry.IsDir() {
			subDirCount++
			if _, err := o.ProcessDir(fp, true); err != nil {
				return 0, err
			}

			continue
		}
		if o.skipcheck(fp) {
			continue
		}

		kind := path.Ext(fp)
		ext := ""
		if kind != "" {
			ext = kind[1:]
		}

		typeDir := o.AddType(ext, fp)
		specialSubDir, err := o.getSpecialSubDirNames(typeDir, ext, fp)
		if err != nil {
			return 0, err
		}
		if err := o.Copy(o.Flags.DstPath, typeDir, specialSubDir, fp); err != nil {
			return 0, err
		}
		processed++
		percentage := float64(processed) / float64(total) * 100 //nolint:mnd
		if processed%max(1, total/20) == 0 && !r {              //nolint:mnd
			slog.Info("progress", "completed", fmt.Sprintf("%.1f%%", percentage))
		}
		extensions = append(extensions, ext)
	}
	extensions = RemoveDuplicateStr(extensions)

	return len(extensions), nil
}

func (o *Operator) Operate() (int, error) {
	switch o.Flags.Async {
	case true:
		return o.AsyncProcessDir(o.Flags.SrcPath, false)
	case false:
		return o.ProcessDir(o.Flags.SrcPath, false)
	}

	return 0, nil
}
