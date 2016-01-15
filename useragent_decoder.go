package useragent

import (
	"errors"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"github.com/ua-parser/uap-go/uaparser"
)

type UserAgentDecoderConfig struct {
	UserAgentFile string `toml:"useragent_file"`
	SourceField   string `toml:"source_field"`
	CacheSize     int    `toml:"cache_size"`
}

type UserAgentDecoder struct {
	UserAgentFile string
	SourceField   string
	CacheSize     int
	parser        *uaparser.Parser
	pConfig       *PipelineConfig
	cache         *lru.TwoQueueCache
}

func (ua *UserAgentDecoder) ConfigStruct() interface{} {
	globals := ua.pConfig.Globals
	return &UserAgentDecoderConfig{
		UserAgentFile: globals.PrependShareDir("ua_regexes.yaml"),
		SourceField:   "",
		CacheSize:     0,
	}
}

func (ua *UserAgentDecoder) SetPipelineConfig(pConfig *PipelineConfig) {
	ua.pConfig = pConfig
}

func (ua *UserAgentDecoder) Init(config interface{}) (err error) {
	conf := config.(*UserAgentDecoderConfig)

	if conf.SourceField == "" {
		return errors.New("`source_field` must be specified")
	}

	ua.SourceField = conf.SourceField

	if ua.parser == nil {
		ua.parser, err = uaparser.New(conf.UserAgentFile)
	}
	if err != nil {
		return fmt.Errorf("Could not open user agent regex file: %s\n")
	}

	if conf.CacheSize > 0 {
		ua.CacheSize = conf.CacheSize
		// We're just using the default cache values
		// defined by the LRU library. Should these be tweakable?
		ua.cache, err = lru.New2Q(ua.CacheSize)
		if err != nil {
			return
		}
	}

	return
}

func (ua *UserAgentDecoder) GetAgent(uaStr string) (*uaparser.Client, bool) {
	var uaClient *uaparser.Client

	if ua.CacheSize > 0 {
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
	var userAgentField, _ = pack.Message.GetFieldValue(ua.SourceField)

	agentStr, ok := userAgentField.(string)
	if !ok {
		// Skip it and move on with next pack
		packs = []*PipelinePack{pack}
		return
	}

	if ua.parser != nil {
		// TODO Track cache hits/misses
		uaClient, _ := ua.GetAgent(agentStr)

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

// TODO Can decoders have metrics like input/output plugins?

func init() {
	RegisterPlugin("UserAgentDecoder", func() interface{} {
		return new(UserAgentDecoder)
	})
}
