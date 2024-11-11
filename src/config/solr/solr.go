package solr

import (
	"fmt"
	"sync"

	"search-courses-api/src/config/envs"

	"github.com/vanng822/go-solr/solr"
	"go.uber.org/zap"
)

var (
	once          sync.Once
	solrInterface *solr.SolrInterface
)

func ConnectSolr(logger *zap.Logger) (*solr.SolrInterface, error) {
	var err error

	once.Do(func() {
		envs := envs.LoadEnvs(".env")
		solrHost := envs.Get("SOLR_HOST")
		solrPort := envs.Get("SOLR_PORT")
		solrCore := envs.Get("SOLR_CORE")

		solrBaseURL := fmt.Sprintf("http://%s:%s/solr", solrHost, solrPort)

		solrInterface, err = solr.NewSolrInterface(solrBaseURL, solrCore)
		if err != nil {
			logger.Error("[SEARCH-API] Error al conectar con Solr", zap.Error(err))
			return
		}

		logger.Info("[SEARCH-API] Conexión a Solr establecida", zap.String("url", solrBaseURL))
	})

	return solrInterface, err
}
