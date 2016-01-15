package useragent

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const AgentQueueLength = 1000

var (
	seedSetup sync.Once

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

func buildRandomAgentOrder(group, upto int) (result []int) {
	seedSetup.Do(func() {
		seed, err := strconv.ParseInt(os.Getenv("SEED"), 10, 64)
		if seed == 0 || err != nil {
			seed = time.Now().UnixNano()
		}
		rand.Seed(seed)
	})

	result = make([]int, upto)
	for i := 0; i < upto; i++ {
		result[i] = rand.Intn(group)
	}

	return
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
	agents := append(badAgents[:], realAgents[:]...)
	input := buildRandomAgentOrder(len(agents), AgentQueueLength)
	inputLen := len(input)

	decoder := buildDecoder(0)

	var idx = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = input[i%inputLen]
		decoder.GetAgent(agents[idx])
	}
}

func BenchmarkAgentParse50PercentCache(b *testing.B) {
	agents := append(badAgents[:], realAgents[:]...)
	agentsLen := len(agents)
	input := buildRandomAgentOrder(agentsLen, AgentQueueLength)
	inputLen := len(input)

	ratioSize := agentsLen / 2
	decoder := buildDecoder(ratioSize)

	var idx = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = input[i%inputLen]
		decoder.GetAgent(agents[idx])
	}
}

func BenchmarkAgentParse90PercentCache(b *testing.B) {
	agents := append(badAgents[:], realAgents[:]...)
	agentsLen := len(agents)
	input := buildRandomAgentOrder(agentsLen, AgentQueueLength)
	inputLen := len(input)

	ratioSize := int(float64(agentsLen) * 0.90)
	decoder := buildDecoder(ratioSize)

	var idx = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = input[i%inputLen]
		decoder.GetAgent(agents[idx])
	}
}

func BenchmarkAgentParse100PercentCache(b *testing.B) {
	agents := append(badAgents[:], realAgents[:]...)
	agentsLen := len(agents)
	input := buildRandomAgentOrder(agentsLen, AgentQueueLength)
	inputLen := len(input)

	decoder := buildDecoder(agentsLen)

	var idx = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = input[i%inputLen]
		decoder.GetAgent(agents[idx])
	}
}

func BenchmarkRealAgentParse(b *testing.B) {
	input := buildRandomAgentOrder(len(realAgents), AgentQueueLength)
	inputLen := len(input)

	decoder := buildDecoder(0)

	var idx = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = input[i%inputLen]
		decoder.GetAgent(realAgents[idx])
	}
}

func BenchmarkBadAgentParse(b *testing.B) {
	input := buildRandomAgentOrder(len(badAgents), AgentQueueLength)
	inputLen := len(input)

	decoder := buildDecoder(0)

	var idx = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = input[i%inputLen]
		decoder.GetAgent(badAgents[idx])
	}
}
