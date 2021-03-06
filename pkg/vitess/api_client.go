package vitess

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/github/freno/pkg/config"
	"vitess.io/vitess/go/vt/proto/topodata"
)

const defaultTimeout = time.Duration(5) * time.Second

// Tablet represents information about a running instance of vttablet.
type Tablet struct {
	MysqlHostname string              `json:"mysql_hostname,omitempty"`
	MysqlPort     int32               `json:"mysql_port,omitempty"`
	Type          topodata.TabletType `json:"type,omitempty"`
}

// IsValidReplica returns a bool reflecting if a tablet type is REPLICA
func (t Tablet) IsValidReplica() bool {
	return t.Type == topodata.TabletType_REPLICA
}

var httpClient = http.Client{
	Timeout: defaultTimeout,
}

func constructAPIURL(settings config.VitessConfigurationSettings) (url string) {
	api := strings.TrimRight(settings.API, "/")
	if !strings.HasSuffix(api, "/api") {
		api = fmt.Sprintf("%s/api", api)
	}
	url = fmt.Sprintf("%s/keyspace/%s/tablets/%s", api, settings.Keyspace, settings.Shard)

	return url
}

// filterReplicaTablets parses a list of tablets, returning replica tablets only
func filterReplicaTablets(tablets []Tablet) (replicas []Tablet) {
	for _, tablet := range tablets {
		if tablet.IsValidReplica() {
			replicas = append(replicas, tablet)
		}
	}
	return replicas
}

// ParseTablets reads from vitess /api/ks_tablets/<keyspace>/[shard] and returns a
// listing (mysql_hostname, mysql_port, type) of REPLICA tablets
func ParseTablets(settings config.VitessConfigurationSettings) (tablets []Tablet, err error) {
	if settings.TimeoutSecs == 0 {
		httpClient.Timeout = defaultTimeout
	} else {
		httpClient.Timeout = time.Duration(settings.TimeoutSecs) * time.Second
	}

	url := constructAPIURL(settings)
	resp, err := httpClient.Get(url)
	if err != nil {
		return tablets, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tablets, err
	}

	err = json.Unmarshal(body, &tablets)
	return filterReplicaTablets(tablets), err
}
