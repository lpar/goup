package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/mholt/archiver"
)

const curext = "tar"
const destdir = "/usr/local"

func semverParse(v string) []int {
	xx := strings.Split(v+".0.0", ".")
	xa, _ := strconv.Atoi(xx[0])
	xb, _ := strconv.Atoi(xx[1])
	xc, _ := strconv.Atoi(xx[2])
	return []int{xa, xb, xc}
}

func semverGreaterThan(xx []int, yy []int) bool {
	if xx[0] > yy[0] {
		return true
	}
	if xx[0] < yy[0] {
		return false
	}
	if xx[1] > yy[1] {
		return true
	}
	if xx[1] < yy[1] {
		return false
	}
	if xx[2] > yy[2] {
		return true
	}
	if xx[2] < yy[2] {
		return false
	}
	return false
}

func getGoDownloads(durl string, curver []int, curos string, curarch string) (string, string) {
	doc, err := htmlquery.LoadURL(durl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't download Go downloads web page: %v\n", err)
		os.Exit(1)
	}

	dlre := regexp.MustCompile(`.*/go([\d\.]+)\.(\w+)-(\w+)\.(\w+)`)

	for _, tr := range htmlquery.Find(doc, "//tr") {
		href := ""
		sha := ""
		a := htmlquery.FindOne(tr, "//a")
		href = htmlquery.SelectAttr(a, "href")
		tt := htmlquery.FindOne(tr, "//tt")
		if tt != nil {
			sha = htmlquery.InnerText(tt)
		}
		if href != "" && sha != "" {
			m := dlre.FindStringSubmatch(href)
			if m != nil {
				ver := m[1]
				os := m[2]
				arch := m[3]
				ext := m[4]
				if os == curos && arch == curarch && ext == curext {
					semver := semverParse(ver)
					if semverGreaterThan(semver, curver) {
						fmt.Printf("Go %s %s %s is available in %s format\n", ver, os, arch, ext)
						return href, sha
					}
				}
			}
		}
	}
	fmt.Println("No more recent Go version found")
	return "", ""
}

func getGoVersion() ([]int, string, string) {
	gobin := filepath.Join(destdir, "go", "bin", "go")
	out, err := exec.Command(gobin, "version").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't run %s to check version: %v\n", gobin, err)
		os.Exit(1)
	}
	gover := regexp.MustCompile(`go version go([\d\.]+) (\w+)/(\w+)`)
	sout := string(out)
	m := gover.FindStringSubmatch(sout)
	if m == nil {
		fmt.Fprintf(os.Stderr, "Can't parse output of `go version` command '%s'\n", sout)
		os.Exit(1)
	}
	v := m[1]
	ver := semverParse(v)
	os := m[2]
	arch := m[3]
	fmt.Printf("You are running Go %s for %s (%s)\n", v, os, arch)
	return ver, os, arch
}

func main() {

	ver, curos, arch := getGoVersion()

	durl, sha := getGoDownloads("https://golang.org/dl/", ver, curos, arch)

	if durl == "" || sha == "" {
		fmt.Println("Nothing to do")
		return
	}

	fmt.Printf("Download and install? ")
	var resp string
	_, err := fmt.Scanln(&resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't read your response: %v\n", err)
		os.Exit(1)
	}
	yn := resp[0]
	if yn != 'y' && yn != 'Y' {
		fmt.Println("Doing nothing")
		return
	}

	downloadAndInstall(durl, sha)

}

func downloadAndInstall(durl string, sha string) {
	req, err := http.NewRequest("GET", durl, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't download Go binary release archive: %v\n", err)
		os.Exit(1)
	}
	u, err := url.Parse(durl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse download URL %s: %v", durl, err)
		os.Exit(1)
	}
	fname := filepath.Base(u.Path)
	fmt.Printf("Downloading %s\n", fname)
	unpacker := archiver.MatchingFormat(fname)
	if unpacker == nil {
		fmt.Fprintf(os.Stderr, "I don't know how to unpack %s\n", fname)
		return
	}

	fmt.Printf("Fetching %s\n", durl)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	bsha := sha256.Sum256(data)
	ssha := hex.EncodeToString(bsha[:])
	if ssha != sha {
		fmt.Fprintf(os.Stderr, "Downloaded data SHA256 mismatch\nexpected %s\ngot %s\n",
			sha, bsha)
		return
	}
	fmt.Println("Downloaded and SHA256 checked")

	godir := filepath.Join(destdir, "go")
	bakgo := filepath.Join(destdir, "go.bak")
	err = os.Rename(godir, bakgo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't rename %s to %s: %v\n", godir, bakgo, err)
		os.Exit(1)
	}

	err = unpacker.Read(bytes.NewReader(data), destdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't unpack archive to %s: %v\n", godir, err)
		os.Exit(1)
	}

	if _, err = os.Stat(godir); err != nil {
		fmt.Fprintf(os.Stderr, "Something went wrong unpacking the archive into %s\n", destdir)
		fmt.Fprintf(os.Stderr, "Old Go version is in %s\n", bakgo)
		os.Exit(1)
	}

	err = os.RemoveAll(bakgo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't remove old Go version in %s: %v", bakgo, err)
		os.Exit(1)
	}

	fmt.Println("Go upgraded successfully")
}
