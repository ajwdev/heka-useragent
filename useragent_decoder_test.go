package useragent

import (
	"testing"
)

var (
	realAgents = [11]string{
		"Manticore 0.4.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
		"Mozilla/5.0 (Linux; Android 5.0.2; LG-V410 Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.83 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 5.0.1; SAMSUNG SCH-I545 4G Build/LRX22C) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/2.1 Chrome/34.0.1847.76 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.106 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.106 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; NP09; NP09; MAAU; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:40.0) Gecko/20100101 Firefox/40.1.0 Waterfox/40.1.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:42.0) Gecko/20100101 Firefox/42.0 Iceweasel/42.0",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0)",
		"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.106 Safari/537.36",
	}

	badAgents = [5]string{
		"flippity floppity floop",
		"-",
		"",
		"Mozilla/5.0 (AndrewOS; CoolPhone 1.2.3; MaibatsuMonstrosity) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.83 Safari/537.36",
		"Mozilla/4.0 (adjkl;hasdfjk;lasdf;ljk3490-73425lndfgsv90usgo hqweru90[)",
	}
)

func buildDecoder(cacheSize int) *UserAgentDecoder {
	decoder := new(UserAgentDecoder)
	decoder.Init(&UserAgentDecoderConfig{
		UserAgentFile: "uap-core/regexes.yaml",
		SourceField:   "doesntmatter",
		CacheSize:     cacheSize,
	})

	return decoder
}

func TestAgentParse(t *testing.T) {
	// NOTE This is intentionally non-exhaustive as the upstream uaparser
	// library has plenty of tests that verify the parsing.
	agentStr := "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"

	// Run once without a cache and once with a cache
	var expected string
	var result string
	for _, v := range []int{0, 10} {
		decoder := buildDecoder(v)
		agent := decoder.GetAgent(agentStr)

		expected = "Mobile Safari"
		if result = agent.UserAgent.Family; expected != result {
			t.Errorf("incorrect user agent family; expected '%s', got '%s'", expected, result)
		}
		expected = "9"
		if result = agent.UserAgent.Major; expected != result {
			t.Errorf("incorrect user agent major; expected '%s', got '%s'", expected, result)
		}
		expected = "0"
		if result = agent.UserAgent.Minor; expected != result {
			t.Errorf("incorrect user agent minor; expected '%s', got '%s'", expected, result)
		}
		expected = "iOS"
		if result = agent.Os.Family; expected != result {
			t.Errorf("incorrect os family; expected '%s', got '%s'", expected, result)
		}
		expected = "9"
		if result = agent.Os.Major; expected != result {
			t.Errorf("incorrect os major; expected '%s', got '%s'", expected, result)
		}
		expected = "1"
		if result = agent.Os.Minor; expected != result {
			t.Errorf("incorrect os minor; expected '%s', got '%s'", expected, result)
		}
		expected = "iPhone"
		if result = agent.Device.Family; expected != result {
			t.Errorf("incorrect device family; expected '%s', got '%s'", expected, result)
		}
	}
}

// TODO Add tests that verify the Heka message pack
// TODO Add tests that verify long version name of the OS (i.e the field ua_os)

func BenchmarkAgentParse(b *testing.B) {
	decoder := buildDecoder()

	agents := append(badAgents[:], realAgents[:]...)
	fixtureLength := len(agents)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder.GetAgent(agents[i%fixtureLength])
	}
}

func BenchmarkRealAgentParse(b *testing.B) {
	decoder := buildDecoder()

	fixtureLength := len(realAgents)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder.GetAgent(realAgents[i%fixtureLength])
	}
}

func BenchmarkBadAgentParse(b *testing.B) {
	decoder := buildDecoder()

	fixtureLength := len(badAgents)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder.GetAgent(badAgents[i%fixtureLength])
	}
}
