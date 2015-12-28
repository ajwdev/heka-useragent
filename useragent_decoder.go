package useragent

import (
	"errors"
	"fmt"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"github.com/ua-parser/uap-go/uaparser"
)

type UserAgentDecoderConfig struct {
	UserAgentFile string `toml:"useragent_file"`
	SourceField   string `toml:"source_field"`
}

type UserAgentDecoder struct {
	UserAgentFile string
	SourceField   string
	parser        *uaparser.Parser
	pConfig       *PipelineConfig
}

func (ua *UserAgentDecoder) ConfigStruct() interface{} {
	globals := ua.pConfig.Globals
	return &UserAgentDecoderConfig{
		UserAgentFile: globals.PrependShareDir("useragent/regexes.yaml"),
		SourceField:   "",
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

	return
}

func (ua *UserAgentDecoder) GetAgent(uaStr string) *uaparser.Client {
	// TODO Add caching for performance because parsing the large number of
	// regexes that we do is not cheap. Also, its common that a single
	// user/browser will make several requests concurrently which means our
	// user agent strings will likely be close to one another in our input
	// stream.
	return ua.parser.Parse(uaStr)
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
		uaClient := ua.GetAgent(agentStr)

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
