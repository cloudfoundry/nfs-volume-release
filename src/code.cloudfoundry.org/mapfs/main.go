// Mounts another directory while mapping uid and gid to a different user.  Extends loopbackfs.

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"code.cloudfoundry.org/goshims/syscallshim"
	"code.cloudfoundry.org/mapfs/mapfs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	debugLogging := flag.Bool("debug", false, "")
	uid := flag.Int64("uid", -1, "")
	gid := flag.Int64("gid", -1, "")
	fsName := flag.String("fsname", "mapfs", "")
	autoCache := flag.Bool("auto_cache", false, "")
	configFile := flag.String("config", "", "configuration file path (optional)")

	flag.Parse()
	if flag.NArg() < 2 || *uid <= 0 || *gid <= 0 {
		fmt.Printf("usage: %s -uid UID -gid GID [-fsname FSNAME] [-auto_cache] [-debug] MOUNTPOINT ORIGINAL\n", path.Base(os.Args[0]))
		fmt.Println("UID and GID must be > 0")
		os.Exit(2)
	}
	if *uid > math.MaxUint32 || *gid > math.MaxUint32 {
		fmt.Printf("usage: %s -uid UID -gid GID [-fsname FSNAME] [-auto_cache] [-debug] MOUNTPOINT ORIGINAL\n", path.Base(os.Args[0]))
		fmt.Printf("UID and GID must be <= %d\n", math.MaxUint32)
		os.Exit(2)
	}
	if *autoCache {
		fmt.Println("warning -- auto_cache flag ignored as it is unsupported in fusermount")
	}
	cfg, err := parseConfigFile(*configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	if cfg.Debug {
		debugLogging = &cfg.Debug
		cleanup := redirectLoggerToFile()
		defer cleanup()
	}
	if cfg.CPUProfile != "" {
		cleanup := startCPUProfile(cfg.CPUProfile)
		defer cleanup()
	}
	if cfg.MemProfile != "" {
		memprofile(cfg.MemProfile)
	}
	if cfg.SoftMemLimit > 0 {
		debug.SetMemoryLimit(cfg.SoftMemLimit)
	}

	orig := flag.Arg(1)
	loopbackfs := pathfs.NewLoopbackFileSystem(orig)
	finalFs := mapfs.NewMapFileSystem(uint32(*uid), uint32(*gid), loopbackfs, orig, &syscallshim.SyscallShim{})

	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}

	pathFs := pathfs.NewPathNodeFs(finalFs, &pathfs.PathNodeFsOptions{})
	conn := nodefs.NewFileSystemConnector(pathFs.Root(), opts)
	mountPoint := flag.Arg(0)
	origAbs, _ := filepath.Abs(orig)
	mOpts := &fuse.MountOptions{
		AllowOther:     true,
		Name:           *fsName,
		FsName:         origAbs,
		Debug:          *debugLogging,
		SingleThreaded: cfg.SingleThreaded,
	}
	state, err := fuse.NewServer(conn.RawFS(), mountPoint, mOpts)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Mounted!")
	state.Serve()
}

// config is the struct used to parse a configuration file
type config struct {
	Debug          bool   `yaml:"debug"`
	SingleThreaded bool   `yaml:"single_threaded"`
	CPUProfile     string `yaml:"cpu_profile"`
	MemProfile     string `yaml:"mem_profile"`
	SoftMemLimit   int64  `yaml:"soft_mem_limit"`
}

// parseConfigFile reads and parses the configuration file
// If no path is specified, but a file is found at the default path then it will
// be read and parsed.
func parseConfigFile(path string) (config, error) {
	const defaultPath = "/var/vcap/jobs/mapfs/config/mapfs.yml"
	if path == "" {
		_, err := os.Stat(defaultPath)
		if err != nil {
			return config{}, nil
		}
		path = defaultPath
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return config{}, fmt.Errorf("could not read config file: %w", err)
	}

	var receiver config
	if err := yaml.Unmarshal(data, &receiver); err != nil {
		return config{}, fmt.Errorf("error parsing config file data: %w", err)
	}

	return receiver, nil
}

// memprofile starts a goroutine that will write a memory profile whenever
// the SIGUSR1 signal is sent to the mapfs process
func memprofile(path string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go func() {
		for range sigs {
			fh, err := os.Create(filepath.Clean(path) + time.Now().Format(".20060102150405"))
			if err != nil {
				log.Printf("could not create memory profile file: %s", err)
				continue
			}

			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(fh); err != nil {
				log.Printf("error writing memory profile: %s", err)
			}
			if err := fh.Close(); err != nil {
				log.Printf("error closing file: %s", err)
			}
		}
	}()
}

// startCPUProfile will start a CPU profile, returning a cleanup callback
func startCPUProfile(path string) func() {
	fh, err := os.Create(filepath.Clean(path))
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(fh); err != nil {
		log.Fatal(err)
	}

	return func() {
		pprof.StopCPUProfile()
		if err := fh.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

// redirectLoggerToFile redirects the standard logger (used by go-fuse) to a log
// file, returning a cleanup callback
func redirectLoggerToFile() func() {
	path := fmt.Sprintf("/var/vcap/sys/log/mapfs/mapfs.%d.log", os.Getpid())
	fh, err := os.Create(filepath.Clean(path))
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(fh)

	return func() {
		if err := fh.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
