package useragent

import (
	"errors"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"github.com/ua-parser/uap-go/uaparser"
	"math"
	"sync"
	"sync/atomic"
)

type UserAgentDecoderConfig struct {
	UserAgentFile string `toml:"useragent_file"`
	SourceField   string `toml:"source_field"`
	CacheSize     int    `toml:"cache_size"`
}

type UserAgentDecoder struct {
	conf *UserAgentDecoderConfig

	processMessageCount int64
	processCacheHit     int64
	processCacheMiss    int64

	parser     *uaparser.Parser
	pConfig    *PipelineConfig
	cache      *lru.TwoQueueCache
	reportLock sync.Mutex
}

func (ua *UserAgentDecoder) ConfigStruct() interface{} {
	globals := ua.pConfig.Globals
	return &UserAgentDecoderConfig{
		UserAgentFile: globals.PrependShareDir("useragent/regexes.yaml"),
		SourceField:   "",
		CacheSize:     0,
	}
}

func (ua *UserAgentDecoder) SetPipelineConfig(pConfig *PipelineConfig) {
	ua.pConfig = pConfig
}

func (ua *UserAgentDecoder) Init(config interface{}) (err error) {
	ua.conf = config.(*UserAgentDecoderConfig)

	if ua.conf.SourceField == "" {
		return errors.New("`source_field` must be specified")
	}

	if ua.parser == nil {
		ua.parser, err = uaparser.New(ua.conf.UserAgentFile)
		if err != nil {
			return fmt.Errorf("Could not open user agent regex file: %s\n")
		}
	}

	if ua.conf.CacheSize > 0 {
		// We're just using the default cache values
		// defined by the LRU library. Should these be tweakable?
		ua.cache, err = lru.New2Q(ua.conf.CacheSize)
		if err != nil {
			return
		}
	}

	return
}

func (ua *UserAgentDecoder) GetAgent(uaStr string) (*uaparser.Client, bool) {
	var uaClient *uaparser.Client

	if ua.conf.CacheSize > 0 {
		if val, ok := ua.cache.Get(uaStr); !ok {
			uaClient = ua.parser.Parse(uaStr)
			// TODO We should track cache evictions
			ua.cache.Add(uaStr, uaClient)
		} else {
			return val.(*uaparser.Client), true
		}
	} else {
		uaClient = ua.parser.Parse(uaStr)
	}

	return uaClient, false
}

func (ua *UserAgentDecoder) Decode(pack *PipelinePack) (packs []*PipelinePack, err error) {
	var userAgentField, _ = pack.Message.GetFieldValue(ua.conf.SourceField)

	agentStr, ok := userAgentField.(string)
	if !ok {
		// Skip it and move on with next pack
		packs = []*PipelinePack{pack}
		return
	}

	if ua.parser != nil {
		uaClient, cacheHit := ua.GetAgent(agentStr)
		if cacheHit {
			atomic.AddInt64(&ua.processCacheHit, 1)
		} else {
			atomic.AddInt64(&ua.processCacheMiss, 1)
		}

		atomic.AddInt64(&ua.processMessageCount, 1)

		var nf *message.Field

		// TODO Handle the err
		// User Agent section
		if uaClient.UserAgent.Family != "" {
			nf, err = message.NewField("ua_name", uaClient.UserAgent.Family, "")
			pack.Message.AddField(nf)
		}
		if uaClient.UserAgent.Major != "" {
			nf, err = message.NewField("ua_major", uaClient.UserAgent.Major, "")
			pack.Message.AddField(nf)
		}
		if uaClient.UserAgent.Minor != "" {
			nf, err = message.NewField("ua_minor", uaClient.UserAgent.Minor, "")
			pack.Message.AddField(nf)
		}
		if uaClient.UserAgent.Patch != "" {
			nf, err = message.NewField("ua_patch", uaClient.UserAgent.Patch, "")
			pack.Message.AddField(nf)
		}
		// OS section
		if uaClient.Os.Family != "" {
			nf, err = message.NewField("ua_os_name", uaClient.Os.Family, "")
			pack.Message.AddField(nf)

			// Add a more readable OS name to match Logstash functionality
			longVersion := uaClient.Os.ToVersionString()
			if longVersion != "" {
				nf, err = message.NewField("ua_os", uaClient.Os.Family+" "+longVersion, "")
				pack.Message.AddField(nf)
			}
		}
		if uaClient.Os.Major != "" {
			nf, err = message.NewField("ua_os_major", uaClient.Os.Major, "")
			pack.Message.AddField(nf)
		}
		if uaClient.Os.Minor != "" {
			nf, err = message.NewField("ua_os_minor", uaClient.Os.Minor, "")
			pack.Message.AddField(nf)
		}
		if uaClient.Os.Patch != "" {
			nf, err = message.NewField("ua_os_patch", uaClient.Os.Patch, "")
			pack.Message.AddField(nf)
		}
		if uaClient.Os.PatchMinor != "" {
			nf, err = message.NewField("ua_os_patch_minor", uaClient.Os.PatchMinor, "")
			pack.Message.AddField(nf)
		}
		// Device section
		if uaClient.Device.Family != "" {
			nf, err = message.NewField("ua_device", uaClient.Device.Family, "")
			pack.Message.AddField(nf)
		}
	}

	packs = []*PipelinePack{pack}
	return
}

func (ua *UserAgentDecoder) ReportMsg(msg *message.Message) error {
	ua.reportLock.Lock()
	defer ua.reportLock.Unlock()

	message.NewInt64Field(msg, "ProcessMessageCount",
		atomic.LoadInt64(&ua.processMessageCount), "count")

	hit := atomic.LoadInt64(&ua.processCacheHit)
	miss := atomic.LoadInt64(&ua.processCacheMiss)
	hitRatio := round((float64(hit) / float64(hit+miss)) * 100)

	message.NewInt64Field(msg, "ProcessCacheHit", hit, "count")
	message.NewInt64Field(msg, "ProcessCacheMiss", miss, "count")
	message.NewInt64Field(msg, "ProcessCacheHitRatio", hitRatio, "percent")
	message.NewInt64Field(msg, "ProcessCacheSize", ua.cache.Len(), "count")
	message.NewInt64Field(msg, "ProcessCacheMaxSize", ua.conf.CacheSize, "count")

	return nil
}

func round(f float64) float64 {
	// Round to 3 decimal places
	return math.Floor((f*1000)+0.5) / 1000
}

func init() {
	RegisterPlugin("UserAgentDecoder", func() interface{} {
		return new(UserAgentDecoder)
	})
}
