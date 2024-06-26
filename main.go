package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/mholt/archiver/v3"
	flag "github.com/spf13/pflag"

	"github.com/lpar/goup/semver"
)

const (
	destDir       = "/usr/local"
	clientTimeout = 5 * 60 * time.Second
)

const (
	dlBase     = "https://dl.google.com/go/"
	dlJSONfeed = "https://golang.org/dl/?mode=json"
)

var (
	unstable  = flag.Bool("unstable", false, "include unstable (beta, rc) versions")
	specOS    = flag.String("os", "", "specify OS (darwin/freebsd/linux)")
	specArch  = flag.String("arch", "", "specify architecture (amd64/arm64/386/ppc64le/s390x/armv6l)")
	force     = flag.Bool("force", false, "force download even if current version up-to-date")
	destGoDir = flag.String("dir", destDir, "destination for go directory (default /usr/local)")
)

// GoDownload represents a download of Go for a specific OS and architecture
type GoDownload struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int    `json:"size"`
	Kind     string `json:"kind"`
}

// GoVersion represents the downloads for a specific release of Go
type GoVersion struct {
	Version string       `json:"version"`
	Stable  bool         `json:"stable"`
	Files   []GoDownload `json:"files"`
}

func pickBestVersion(targetOS string, targetArch string) (*GoVersion, *GoDownload, error) {
	var bestVersion *GoVersion
	var bestDownload *GoDownload
	client := http.Client{
		Timeout: clientTimeout,
	}
	req, err := http.NewRequest(http.MethodGet, dlJSONfeed, nil)
	if err != nil {
		return bestVersion, bestDownload, fmt.Errorf("can't create HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return bestVersion, bestDownload, fmt.Errorf("request failed: %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bestVersion, bestDownload, fmt.Errorf("can't read body: %w", err)
	}
	var availableVersions []GoVersion
	err = json.Unmarshal(body, &availableVersions)
	if err != nil {
		return bestVersion, bestDownload, fmt.Errorf("can't parse JSON: %w", err)
	}
	sort.Slice(availableVersions, func(i, j int) bool {
		v1 := semver.NewSemVer(availableVersions[i].Version)
		v2 := semver.NewSemVer(availableVersions[j].Version)
		return v1.GreaterThan(v2)
	})
	for _, version := range availableVersions {
		if version.Stable || *unstable {
			download, err := pickBestFile(version, targetOS, targetArch)
			if err == nil {
				bestVersion = &version
				bestDownload = download
				return bestVersion, bestDownload, nil
			}
		}
	}
	return bestVersion, bestDownload, fmt.Errorf("no availableVersions found for %s/%s", targetOS, targetArch)
}

func pickBestFile(gv GoVersion, targetOS string, targetArch string) (*GoDownload, error) {
	for _, file := range gv.Files {
		if file.Arch == targetArch && file.OS == targetOS {
			return &file, nil
		}
	}
	return nil, fmt.Errorf("no viable download for %s (%s) in %s", targetOS, targetArch, gv.Version)
}

func getCurrentGoVersion() (string, string, string, error) {
	var ver string
	var opsys string
	var arch string
	gobin := filepath.Join(*destGoDir, "go", "bin", "go")
	out, err := exec.Command(gobin, "version").Output()
	if err != nil {
		out, err = exec.Command("go", "version").Output()
		if err != nil {
			return ver, opsys, arch, fmt.Errorf("can't run %s to check version: %w", gobin, err)
		}
	}
	gover := regexp.MustCompile(`go version go([\d.]+) (\w+)/(\w+)`)
	sout := string(out)
	m := gover.FindStringSubmatch(sout)
	if m == nil {
		return ver, opsys, arch, fmt.Errorf("can't parse output of `go version` command ('%s')", sout)
	}
	ver = m[1]
	opsys = m[2]
	arch = m[3]
	return ver, opsys, arch, nil
}

func main() {
	flag.Parse()

	currentVersion, targetOS, targetArch, err := getCurrentGoVersion()
	if err != nil {
		fmt.Printf("Can't determine what version of Go you are running: %v\n", err)
	} else {
		fmt.Printf("You are running Go %s for %s (%s)\n", currentVersion, targetOS, targetArch)
	}
	if *specArch != "" {
		targetArch = *specArch
		fmt.Printf("Using architecture %s as specified on command line", targetArch)
	}
	if *specOS != "" {
		targetOS = *specOS
		fmt.Printf("Using OS %s as specified on command line", targetOS)
	}

	if targetOS == "" || targetArch == "" {
		fmt.Println("Can't proceed without knowing the OS and architecture you require (see --help for how to specify)")
		return
	}

	fmt.Printf("Checking available Go versions for %s (%s)\n", targetOS, targetArch)

	newVersion, newDownload, err := pickBestVersion(targetOS, targetArch)
	if err != nil {
		fmt.Printf("Couldn't check for new versions: %v", err)
		return
	}
	newSemVer := semver.NewSemVer(newVersion.Version)

	fmt.Printf("Found Go %s for %s (%s)\n", newSemVer.String(), newDownload.OS, newDownload.Arch)

	if currentVersion != "" {
		curSemVer := semver.NewSemVer(currentVersion)
		if !newSemVer.GreaterThan(curSemVer) {
			fmt.Println("Current version is up to date")
			if !*force {
				return
			}
		}
	}

	fmt.Printf("Download and install Go %s for %s (%s)? ", newSemVer.String(), newDownload.OS, newDownload.Arch)

	var resp string
	_, err = fmt.Scanln(&resp)
	if err != nil {
		fmt.Printf("Can't read your response: %v\n", err)
		os.Exit(1)
	}

	yn := resp[0]
	if yn != 'y' && yn != 'Y' {
		fmt.Println("Doing nothing")
		return
	}

	err = downloadAndInstall(newDownload)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// Download a file and return the temporary filename where it's stored and its SHA256 checksum.
// If the download fails for any reason, attempt to clean up the temporary file.
func downloadFile(dl *GoDownload) (string, string, error) {
	var tmpfile string
	var ssha string

	cleanup := func() {
		clerr := os.Remove(tmpfile)
		if clerr != nil {
			fmt.Fprintf(os.Stderr, "error cleaning up temporary file %s: %v", tmpfile, clerr)
		}
	}

	srcurl := dlBase + dl.Filename
	u, err := url.Parse(srcurl)
	if err != nil {
		return tmpfile, ssha, fmt.Errorf("can't parse download URL %s: %w", srcurl, err)
	}
	fname := filepath.Base(u.Path)
	fmt.Printf("Downloading %s\n", fname)
	tmpfile = path.Join(os.TempDir(), fname)
	fp, err := os.OpenFile(tmpfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("can't open temporary file %s: %w", tmpfile, err)
	}
	client := http.Client{
		Timeout: clientTimeout,
	}
	req, err := http.NewRequest(http.MethodGet, srcurl, nil)
	if err != nil {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("can't create HTTP request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("download request gave HTTP %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("can't read body: %w", err)
	}
	nbytes, err := fp.Write(body)
	if err != nil {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("can't write downloaded data to %s: %w", tmpfile, err)
	}
	if nbytes != dl.Size {
		return tmpfile, ssha, fmt.Errorf("wrong download size, expected %d bytes, got %d", dl.Size, nbytes)
	}
	err = fp.Close()
	if err != nil {
		cleanup()
		return tmpfile, ssha, fmt.Errorf("can't close %s: %w", tmpfile, err)
	}
	bsha := sha256.Sum256(body)
	ssha = hex.EncodeToString(bsha[:])
	return tmpfile, ssha, nil
}

func fixPermissions(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		perms := info.Mode()
		perms = (perms & 0o777) | 0o444
		if info.IsDir() {
			perms = perms | 0o111
		}
		cherr := os.Chmod(path, perms)
		if cherr != nil {
			return fmt.Errorf("can't chmod %###o %s", perms, path)
		}
		return err
	})
}

func downloadAndInstall(dl *GoDownload) error {
	tmpfile, shasum, err := downloadFile(dl)
	if err != nil {
		return fmt.Errorf("temporary file download failed: %w", err)
	}
	fmt.Printf("Temporary file downloaded successfully into %v\n", tmpfile)
	if shasum != dl.SHA256 {
		return fmt.Errorf("bad checksum, expected %s got %s", dl.SHA256, shasum)
	}
	fmt.Println("Downloaded and SHA256 verified")
	godir := filepath.Join(*destGoDir, "go")
	bakgo := filepath.Join(*destGoDir, "go.bak")
	if _, err = os.Stat(godir); !os.IsNotExist(err) {
		err = os.Rename(godir, bakgo)
		if err != nil {
			return fmt.Errorf("can't rename %s to %s: %w", godir, bakgo, err)
		}
	}
	if err = archiver.Unarchive(tmpfile, *destGoDir); err != nil {
		return fmt.Errorf("can't unpack %s to %s: %w", tmpfile, godir, err)
	}
	if _, err = os.Stat(godir); err != nil {
		return fmt.Errorf("problem unpacking to %s, old go version is in %s", godir, bakgo)
	}
	if err = os.Remove(tmpfile); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't remove temporary file %s: %v", tmpfile, err)
	}
	if err = os.RemoveAll(bakgo); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't remove old Go version in %s: %v", bakgo, err)
	}
	if err = fixPermissions(godir); err != nil {
		return fmt.Errorf("error installing: %w", err)
	}
	fmt.Println("Go upgraded successfully")
	return nil
}
