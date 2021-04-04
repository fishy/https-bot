package similarity_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"testing/quick"

	"github.com/fishy/https-bot/similarity"
)

var sizes = []int{16, 256, 512, 1024, 5120, 10240}

func TestSimilarity(t *testing.T) {
	for _, c := range []struct {
		a, b     string
		expected int
	}{
		{
			a:        "",
			b:        "",
			expected: 0,
		},
		{
			a:        "",
			b:        "foo",
			expected: 0,
		},
		{
			a:        "foo",
			b:        "",
			expected: 0,
		},
		{
			a:        "abc",
			b:        "abc",
			expected: 3,
		},
		{
			a:        "abc",
			b:        "def",
			expected: 0,
		},
		{
			a:        "abcdef",
			b:        "abcfoodef",
			expected: 6,
		},
		{
			a:        "abcdef",
			b:        "abc",
			expected: 3,
		},
		{
			a:        "def",
			b:        "abcdef",
			expected: 3,
		},
		{
			// This is a case an optimization in LCS can be implemented incorrectly.
			// At first LCS would find that a and b has "abc" in common,
			// if an optimization skipped the iteration in a directly to index 3 next,
			// it would not find the longer common substrng of "bcde".
			a:        "abcde",
			b:        "bcdeabc",
			expected: 4,
		},
	} {
		t.Run(fmt.Sprintf("%s-%s", c.a, c.b), func(t *testing.T) {
			actual := similarity.Similarity([]byte(c.a), []byte(c.b))
			if actual != c.expected {
				t.Errorf(
					"Expected Similarity(%q, %q) to return %d, got %d",
					c.a,
					c.b,
					c.expected,
					actual,
				)
			}
		})
	}
}

func generateContent(tb testing.TB, r io.Reader, size int) []byte {
	tb.Helper()

	var n int
	var err error
	buf := make([]byte, size)
	n, err = r.Read(buf)
	if err != nil || n != size {
		tb.Fatalf("Read %d of %d, err: %v", n, size, err)
	}
	return buf
}

func TestSimilarityQuick(t *testing.T) {
	for _, size := range sizes {
		t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
			t.Run("identical", func(t *testing.T) {
				f := func() bool {
					buf := generateContent(t, rand.Reader, size)
					if actual := similarity.Similarity(buf, buf); actual != size {
						t.Errorf(
							"Expected %d, got %d\n% 02x",
							size,
							actual,
							buf,
						)
					}
					return !t.Failed()
				}
				if err := quick.Check(f, nil); err != nil {
					t.Error(err)
				}
			})

			const scaleBase = float64(128)
			var scale float64
			if s := float64(size); s > scaleBase {
				scale = scaleBase / s
			}
			cfg := &quick.Config{
				MaxCountScale: scale,
			}

			t.Run("flip-one", func(t *testing.T) {
				f := func(index uint) bool {
					a := generateContent(t, rand.Reader, size)
					b := make([]byte, size)
					copy(b, a)
					index = index % uint(size)
					b[index] = b[index] ^ 0xff
					expected := size - 1
					if actual := similarity.Similarity(a, b); actual != expected {
						t.Errorf(
							"Expected %d, got %d\n% 02x\n% 02x",
							expected,
							actual,
							a,
							b,
						)
					}
					return !t.Failed()
				}
				if err := quick.Check(f, cfg); err != nil {
					t.Error(err)
				}
			})

			const target = 0.4
			t.Run("different", func(t *testing.T) {
				f := func() bool {
					a := generateContent(t, rand.Reader, size)
					b := generateContent(t, rand.Reader, size)
					sim := similarity.Similarity(a, b)
					if float64(sim) > float64(size)*target {
						t.Errorf(
							"Too similar: %d/%d\n% 02x\n% 02x",
							sim,
							size,
							a,
							b,
						)
					}
					return !t.Failed()
				}
				if err := quick.Check(f, cfg); err != nil {
					t.Error(err)
				}
			})
		})
	}
}

func BenchmarkSimilarity(b *testing.B) {
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			aa := generateContent(b, rand.Reader, size)
			bb := generateContent(b, rand.Reader, size)
			if actual := similarity.Similarity(aa, aa); actual != size {
				b.Fatalf(
					"Expected %d, got %d\n% 02x",
					size,
					actual,
					aa,
				)
			}
			if actual := similarity.Similarity(bb, bb); actual != size {
				b.Fatalf(
					"Expected %d, got %d\n% 02x",
					size,
					actual,
					bb,
				)
			}
			b.Logf(
				"aa vs. bb: %d/%d",
				similarity.Similarity(aa, bb),
				size,
			)
			b.ResetTimer()

			b.Run("identical", func(b *testing.B) {
				b.ReportAllocs()

				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						similarity.Similarity(aa, aa)
					}
				})
			})

			b.Run("different", func(b *testing.B) {
				b.ReportAllocs()

				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						similarity.Similarity(aa, bb)
					}
				})
			})

			b.Run("flip-one", func(b *testing.B) {
				cc := make([]byte, size)
				copy(cc, aa)
				index := size / 2
				cc[index] = cc[index] ^ 0xff
				expected := size - 1
				if actual := similarity.Similarity(aa, cc); actual != expected {
					b.Fatalf(
						"Expected %d, got %d\n% 02x\n% 02x",
						expected,
						actual,
						aa,
						cc,
					)
				}
				b.ResetTimer()
				b.ReportAllocs()

				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						similarity.Similarity(aa, cc)
					}
				})
			})
		})
	}
}
