package ipip

import (
	"testing"
)

func TestIpFind(t *testing.T) {
	if err := Init(); err != nil {
		panic(err)
	}
	var tests = []struct {
		ip       string
		country  string
		province string
	}{
		{"180.168.197.84", "中国", "上海"},
		{"218.28.191.40", "中国", "河南"},
		{"123.125.71.38", "中国", "湖北"},
	}

	for _, test := range tests {
		if res, err := Find(test.ip); err != nil {
			if res.Country != test.country || res.Province != test.province {
				t.Errorf("Find(%q) = %v", test.ip, res)
			}
		}
	}
}

func BenchmarkIpFind(b *testing.B) {
	if err := Init(); err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		Find("180.168.197.84")
	}
}
