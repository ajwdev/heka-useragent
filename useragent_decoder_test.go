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

func buildDecoder() *UserAgentDecoder {
	decoder := new(UserAgentDecoder)
	decoder.Init(&UserAgentDecoderConfig{
		UserAgentFile: "uap-core/regexes.yaml",
		SourceField:   "doesntmatter",
	})

	return decoder
}

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
