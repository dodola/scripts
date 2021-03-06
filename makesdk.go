///bin/true ; exec /usr/bin/env go run "$0" "$@"
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var output = flag.String("output", "fuchsia-sdk.tgz", "Output name")
var toolchain = flag.Bool("toolchain", false, "Include toolchain")
var toolchainLibs = flag.Bool("toolchain-lib", true, "Include toolchain libraries in SDK. Typically used when --toolchain is false")
var sysroot = flag.Bool("sysroot", true, "Include sysroot")
var kernelImg = flag.Bool("kernel-img", true, "Include kernel image")
var kernelDebugObjs = flag.Bool("kernel-dbg", true, "Include kernel objects with debug symbols")
var qemu = flag.Bool("qemu", true, "Include QEMU binary")
var tools = flag.Bool("tools", true, "Include additional tools")
var verbose = flag.Bool("v", false, "Verbose output")
var dryRun = flag.Bool("n", false, "Dry run - print what would happen but don't actually do it")

type compType int

const (
	dir compType = iota
	file
	custom
)

type component struct {
	flag      *bool  // Flag controlling whether this component should be included
	srcPrefix string // Source path prefix relative to the fuchsia root
	dstPrefix string // Destination path prefix relative to the SDK root
	t         compType
	f         func(src, dst string) // When t is 'custom', function to run to copy
}

var components = []component{
	{
		toolchain,
		"buildtools/toolchain",
		"toolchain",
		dir,
		nil,
	},
	{
		toolchainLibs,
		"buildtools/toolchain/clang+llvm-x86_64-linux/lib/clang/5.0.0/lib/fuchsia",
		"toolchain_libs/clang/5.0.0/lib/fuchsia",
		dir,
		nil,
	},
	{
		sysroot,
		"out/sysroot",
		"sysroot",
		dir,
		nil,
	},
	{
		kernelImg,
		"out/build-magenta/build-magenta-pc-x86-64/magenta.bin",
		"kernel/magenta.bin",
		file,
		nil,
	},
	{
		kernelDebugObjs,
		"out/build-magenta/build-magenta-pc-x86-64",
		"kernel/debug",
		custom,
		copyKernelDebugObjs,
	},
	{
		qemu,
		"buildtools/qemu",
		"qemu",
		dir,
		nil,
	},
	{
		tools,
		"out/build-magenta/tools",
		"tools",
		dir,
		nil,
	},
}

func copyKernelDebugObjs(src, dstPrefix string) {
	// The kernel debug information lives in many .elf files in the out directory
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".elf" {
			dst := filepath.Join(dstPrefix, path[len(src):])
			mkdir(filepath.Dir(dst))
			cp(path, dst)
		}
		return nil
	})
	cp(filepath.Join(src, "ids.txt"), filepath.Join(dstPrefix, "ids.txt"))
}

func mkdir(dir string) {
	if *verbose || *dryRun {
		fmt.Println("Making directory", dir)
	}
	if *dryRun {
		return
	}
	_, err := exec.Command("mkdir", "-p", dir).Output()
	if err != nil {
		log.Fatal("could not create directory", dir)
	}
}

func cp(args ...string) {
	if *verbose || *dryRun {
		fmt.Println("Copying", args)
	}
	if *dryRun {
		return
	}
	out, err := exec.Command("cp", args...).CombinedOutput()
	if err != nil {
		log.Fatal("cp failed with output", string(out), "error", err)
	}
}

func copyFile(src, dst string) {
	mkdir(filepath.Dir(dst))
	cp(src, dst)
}

func copyDir(src, dst string) {
	mkdir(filepath.Dir(dst))
	cp("-r", src, dst)
}

func tar(src, dst string) {
	if *verbose || *dryRun {
		fmt.Println("Archiving", src, "to", dst)
	}
	if *dryRun {
		return
	}
	out, err := exec.Command("tar", "cvzf", dst, "-C", src, ".").Output()
	if err != nil {
		log.Fatal("tar failed with output", string(out), "error", err)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ./makesdk.go [flags] /path/to/fuchsia/root

This script creates a Fuchsia SDK containing the specified features and places it into a tarball.
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	fuchsiaRoot := flag.Arg(0)
	if _, err := os.Stat(fuchsiaRoot); os.IsNotExist(err) {
		flag.Usage()
		log.Fatalf("Fuchsia root not found at \"%v\"\n", fuchsiaRoot)
	}
	tmpSdk, err := ioutil.TempDir("", "fuchsia-sdk")
	if err != nil {
		log.Fatal("Could not create temporary directory: ", err)
	}
	defer os.RemoveAll(tmpSdk)

	for _, c := range components {
		if *c.flag {
			src := filepath.Join(fuchsiaRoot, c.srcPrefix)
			dst := filepath.Join(tmpSdk, c.dstPrefix)
			switch c.t {
			case dir:
				copyDir(src, dst)
			case file:
				copyFile(src, dst)
			case custom:
				c.f(src, dst)
			}
		}
	}
	tar(tmpSdk, *output)
}
