package melody

import (
	"encoding/json"
	"fmt"
	"github.com/mdy/melody/resolver/types"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/url"
	"strconv"
)

const (
	// melodyAPI GraphURL endpoint
	melodyURL = "https://api.melody.sh/graphql"

	// melodyAPI release download URL from Name + Revision
	melodyReleaseURL = "https://api.melody.sh/%s/-/%s/tgz"
)

func (p *Melody) fetchSpecs(query *packageQuery) ([]types.Specification, error) {
	// Populate arguments into query and send it to Melody-API
	resp, err := p.client.PostForm(melodyURL, url.Values{"query": {query.GqlString()}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := p.responseError(resp); err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Dig into JSON path "data.package"
	respJSON := struct {
		Data struct {
			Package map[string]json.RawMessage
		}
	}{}

	if err := json.Unmarshal(raw, &respJSON); err != nil {
		return nil, err
	}

	// Let's unmarshall everything one by one
	mSpecs, pkgJSON := []*melodySpec{}, respJSON.Data.Package
	if query.allTagged {
		if err := json.Unmarshal(pkgJSON["versionList"], &mSpecs); err != nil {
			return nil, errors.Wrap(err, "Could not parse JSON versionList")
		}
		delete(pkgJSON, "versionList")
	}

	// Convert []*melodySpec to []resolver.Specification
	specs := make([]types.Specification, len(mSpecs))
	for i, s := range mSpecs {
		specs[i] = s
	}

	// Unmarshall all the individual versions
	for count := 0; ; count++ {
		raw, ok := pkgJSON["v"+strconv.Itoa(count)]
		if !ok || raw == nil {
			break
		}

		spec := &melodySpec{}
		if err := json.Unmarshal(raw, spec); err != nil {
			return nil, errors.Wrap(err, "Could not parse version JSON")
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

type packageQuery struct {
	name      string
	allTagged bool
	revisions []string
	versions  []string
}

func (q *packageQuery) GqlString() string {
	query, vCount := "", 0

	if q.allTagged {
		query += gqlAllTaggedVersions
	}

	for _, ver := range q.versions {
		query += fmt.Sprintf(gqlVersionByName, vCount, strconv.QuoteToASCII(ver))
		vCount++
	}

	for _, rev := range q.revisions {
		query += fmt.Sprintf(gqlVersionByRev, vCount, strconv.QuoteToASCII(rev))
		vCount++
	}

	return fmt.Sprintf(gqlPackageQuery, strconv.QuoteToASCII(q.name), query)
}

const (
	gqlAllTaggedVersions = "versionList { ...VersionInfo }\n"
	gqlVersionByName     = "v%d: version(version:%s) { ...VersionInfo }\n"
	gqlVersionByRev      = "v%d: version(revision:%s) { ...VersionInfo }\n"
	gqlPackageQuery      = `
    query PackageQuery {
      package(name:%s) {
        %s
      }
    }
    fragment VersionInfo on Version {
      name, version,
      release { name, version, revision, url },
      dependencyList(scope:BUILD) { name, versionRange }
    }
  `
)
