package melody

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/mdy/melody/resolver"
	"github.com/mdy/melody/resolver/types"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxParallelInstalls = 5

type Melody struct {
	resolver.BaseProvider
	sessionID string
	client    *http.Client
	base      *resolver.Graph
	cache     *Cache
}

func New(base *resolver.Graph) *Melody {
	source := &Melody{base: base, sessionID: uuid.NewV4().String()}
	source.cache = NewCache(source.fetchAvailableSpecs)
	source.client = &http.Client{Transport: source}
	return source
}

// Look for specifications that match passed-in dependency (name + requirement)
func (p *Melody) SearchFor(req types.Requirement) []types.Specification {
	// Looking for a melodyRelease gets you that melodyRelease
	if mSpec, isRelease := req.(*melodyRelease); isRelease {
		return []types.Specification{mSpec}
	}

	// Let's check the cache for matches first
	availableSpecs, err := p.cache.Fetch(req.Name())
	if err != nil {
		log.Fatalf("Error fetching specifications for %s: %s", req.Name(), err)
		return nil // TODO: Let's have an Err() accessor for Provider!
	}

	// Filter specifications for matches
	specs := []types.Specification{}
	for _, spec := range availableSpecs {
		if p.IsRequirementSatisfiedBy(req, nil, spec) {
			specs = append(specs, spec)
		}
	}

	// We're done, if we have matches
	dep, ok := req.(*melodyRequirement)
	if len(specs) != 0 || !ok {
		return specs
	}

	// Let's try to fetch a specific non-available spec
	pQuery := packageQuery{name: dep.Name(), allTagged: false}
	if strings.HasPrefix(dep.RangeStr, "#") {
		pQuery.revisions = []string{dep.RangeStr[1:]}
	} else {
		pQuery.versions = []string{dep.RangeStr}
	}

	availableSpecs, err = p.fetchSpecs(&pQuery)
	if err != nil {
		log.Fatalf("Error fetching specifications for %s: %s", req.Name(), err)
		return nil // TODO: Let's have an Err() accessor for Provider!
	}

	// Include these in our local cache
	p.cache.Append(dep.Name(), availableSpecs)

	// Filter specifications for matches
	for _, spec := range availableSpecs {
		if p.IsRequirementSatisfiedBy(req, nil, spec) {
			specs = append(specs, spec)
		}
	}

	return specs
}

func (p *Melody) DependenciesFor(spec types.Specification) types.Requirements {
	return spec.Requirements()
}

func (p *Melody) IsRequirementSatisfiedBy(d types.Requirement, _ *resolver.Graph, spec types.Specification) bool {
	ok, err := d.SatisfiedBy(spec)
	if err != nil {
		panic(fmt.Sprintf("Cannot check version %s vs. %s: %s", spec.Version(), d, err.Error()))
	}
	return ok
}

// Filter and install Release specs into specified vendor directory
func (p *Melody) InstallToDir(rootDir string, specs []types.Specification) (err error) {
	var g errgroup.Group
	relChan := make(chan *melodyRelease)

	g.Go(func() error {
		defer close(relChan)
		for _, spec := range specs {
			if release, ok := spec.(*melodyRelease); ok {
				relChan <- release
			}
		}
		return nil
	})

	// Start a fixed number of goroutines to read and digest files.
	for i := 0; i < maxParallelInstalls; i++ {
		g.Go(func() error {
			for release := range relChan {
				err := p.installRelease(rootDir, release)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	return g.Wait()
}

// Install a Release into a specified directory
func (p *Melody) installRelease(rootDir string, release *melodyRelease) error {
	relName := release.NameStr
	if relName == "" {
		return fmt.Errorf("No release Name for %s", relName)
	} else if release.URL == "" {
		return fmt.Errorf("No download URL for %s", relName)
	}

	target := filepath.Join(rootDir, release.InstallPath())
	relDesc := relName + " " + release.Version()
	log.Info("----> RELEASE: ", relName, " to ", target)

	// Manage existing version (keep or remove/replace)
	versionFile := filepath.Join(target, ".melody.ver")
	if stat, err := os.Stat(target); err == nil && stat.IsDir() {
		version, _ := ioutil.ReadFile(versionFile)
		if string(version) == release.Version() {
			fmt.Printf("♫ Using %s\n", relDesc)
			return nil
		}

		log.Infof("Replacing existing release: %s", relDesc)
		os.RemoveAll(target)
	}

	resp, err := p.client.Get(release.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := p.responseError(resp); err != nil {
		return err
	}

	if l := resp.ContentLength; l >= 0 {
		relDesc += " (" + humanize.Bytes(uint64(l)) + ")"
	}

	fmt.Printf("♫ Installing %s\n", relDesc)
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}

	err = os.MkdirAll(target, 0755)
	if err != nil {
		return err
	}

	for tarReader := tar.NewReader(gzReader); ; {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		if i := strings.IndexRune(header.Name, os.PathSeparator); i >= 0 {
			path = filepath.Join(target, header.Name[i+1:])
		}

		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := writeFileFromTar(path, info, tarReader); err != nil {
			return err
		}
	}

	// Commit version file
	ioutil.WriteFile(versionFile, []byte(release.Version()), 0644)
	return nil
}

func writeFileFromTar(path string, info os.FileInfo, reader io.Reader) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = io.Copy(file, reader); err != nil {
		return err
	}
	return nil
}

// Specification caching helpers
func (p *Melody) fetchAvailableSpecs(name string) ([]types.Specification, error) {
	// Query for tagged versions and latest HEAD revision
	pQuery := packageQuery{name: name, allTagged: true}
	pQuery.revisions = append(pQuery.revisions, "HEAD")

	// Existing specs in Lockfile don't have requirements, so we
	// have to make a query to explicitly retrieve full spec :-/
	if p.base != nil {
		if bare := p.base.PayloadFor(name); bare != nil {
			pQuery.versions = append(pQuery.versions, bare.Version())
		}
	}

	return p.fetchSpecs(&pQuery)
}

// Implement http.RoundTripper so we can use our own client
func (p *Melody) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Melody-Session-ID", p.sessionID)
	return http.DefaultTransport.RoundTrip(req)
}

// Read error body returned from Melody server
func (p *Melody) responseError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	data := struct{ Error string }{}
	if err := json.Unmarshal(raw, &data); err != nil {
		data.Error = "Response: " + string(raw)
	}

	return fmt.Errorf("Server error: %s", data.Error)
}
