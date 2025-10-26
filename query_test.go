package jqx

import (
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	query := new(State).Compile(`range(.)*2+1 | tostring`)
	got := slices.Collect(query(3))

	assertString(t, got, `[1 3 5]`)

	i := 0
	for v := range query(100) {
		if len(v.(string)) != 1 {
			break
		}
		i++
	}
	assertEqual(t, i, 5)
}
func TestSmoke(t *testing.T) {
	query := new(State).Compile(`_itertest`)
	assertString(t, slices.Collect(query("hello.")), `[hello. ello. llo. lo. o. .]`)
	assertString(t, slices.Collect(query(5)), `[4 3 2 1 0]`)
}
func TestError(t *testing.T) {
	err := func(code string) (rt error) {
		defer catch[error](&rt)
		new(State).Compile(constString(code))
		return nil
	}

	assertEqual(t, err("."), nil)
	assertEqual(t, err("true"), nil)
	assertEqual(t, err("false"), nil)

	assertEqual(t, err("|") != nil, true)
	assertEqual(t, err("maybe") != nil, true)
}
func TestState(t *testing.T) {
	state := State{}
	query := state.Compile(`
		fromjson | to_entries[] |
		snapshot(.key;.value+.value) | .key+.key
	`)

	keys := slices.Collect(query(`{"a":"aa", "c":[3], "q":{"e":10}, "r":5}`))

	assertEqual(t, len(keys), 4)
	assertEqual(t, len(state.Files), 4)

	state = State{Globals: map[string]any{
		"k": keys,
		"v": state.Files,
	}}
	query = state.Compile(`$k[],$v[] | tostring`)

	got := slices.Collect(query(nil))
	assertString(t, got, `[aa cc qq rr aaaa [3,3] {"e":10} 10]`)
}

var (
	newlineRe = regexp.MustCompile(`\s*\n\s*`)
	spacesRe  = regexp.MustCompile(`\s+`)
)

func checkPagetrim(t *testing.T, s string) string {
	t.Helper()
	rt := pagetrim(s, nil).(string)
	assertEqual(t, pagetrim(rt, nil).(string), rt)
	assertEqual(t, len(rt) > len(s), false)

	assertEqual(t, strings.HasPrefix(rt, "\n"), false)
	assertEqual(t, strings.HasSuffix(rt, "\n"), false)
	assertEqual(t, strings.Contains(rt, "\n\n\n"), false)
	assertEqual(t, strings.Contains(rt, "\n "), false)
	assertEqual(t, strings.Contains(rt, " \n"), false)

	assertEqual(t, strings.Contains(rt, "\n\n"), strings.Contains(strings.TrimSpace(s), "\n\n"))

	assertEqual(t, spacesRe.ReplaceAllString(rt, ""), spacesRe.ReplaceAllString(s, ""))
	assertEqual(t, newlineRe.ReplaceAllString(rt, ""), newlineRe.ReplaceAllString("\n"+s+"\n", ""))
	return rt
}
func TestTrim(t *testing.T) {
	for i := range 10 {
		s := strings.Repeat(" ", i)
		assertEqual(t, checkPagetrim(t, s), "")
		s = strings.Repeat("\n", i)
		assertEqual(t, checkPagetrim(t, s), "")
	}

	const L = 16
	lines := strings.Repeat("\n", L)
	for y := range L {
		for x := range L {
			bt := []byte(lines)
			bt[x] = 'a'
			assertEqual(t, checkPagetrim(t, string(bt)), "a")

			bt[y] = 'a'
			trimmed := checkPagetrim(t, string(bt))

			dist := y - x
			if dist < 0 {
				dist = -dist
			}
			dist++
			switch dist {
			case 1:
				assertEqual(t, trimmed, "a")
			case 2:
				assertEqual(t, trimmed, "aa")
			case 3:
				assertEqual(t, trimmed, "a\na")
			default:
				assertEqual(t, trimmed, "a\n\na")
			}
		}
	}

	for n := range 10 {
		want := "a\n" + strings.Repeat("a\n\n", n) + "a"
		s := ""
		for i := range n + 1 {
			s += "a" + strings.Repeat("\n", i+1)
		}
		s += "a"

		assertEqual(t, checkPagetrim(t, s), want)
	}
}

func TestHash(t *testing.T) {
	query := new(State).Compile(`md5`)
	assertString(t, slices.Collect(query("")), `[d41d8cd98f00b204e9800998ecf8427e]`)
	assertString(t, slices.Collect(query("\n")), `[68b329da9893e34099c7d8ad5cb9c940]`)
	assertString(t, slices.Collect(query("hi")), `[49f68a5c8493ec2c0bf489821c21fc3b]`)

	query = new(State).Compile(`sha1`)
	assertString(t, slices.Collect(query("")), `[da39a3ee5e6b4b0d3255bfef95601890afd80709]`)
	assertString(t, slices.Collect(query("\n")), `[adc83b19e793491b1c6ea0fd8b46cd9f32e592fc]`)
	assertString(t, slices.Collect(query("hi")), `[c22b5f9178342609428d6f51b2c5af4c0bde6a42]`)
	query = new(State).Compile(`sha256`)
	assertString(t, slices.Collect(query("")), `[e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]`)
	assertString(t, slices.Collect(query("\n")), `[01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b]`)
	assertString(t, slices.Collect(query("hi")), `[8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4]`)
	query = new(State).Compile(`sha512`)
	assertString(t, slices.Collect(query("")), `[cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e]`)
	assertString(t, slices.Collect(query("\n")), `[be688838ca8686e5c90689bf2ab585cef1137c999b48c70b92f67a5c34dc15697b5d11c982ed6d71be1e1e7f7b4e0733884aa97c3f7a339a8ed03577cf74be09]`)
	assertString(t, slices.Collect(query("hi")), `[150a14ed5bea6cc731cf86c41566ac427a8db48ef1b9fd626664b3bfbb99071fa4c922f33dde38719b8c8354e2b7ab9d77e0e67fc12843920a712e73d558e197]`)
}
func TestShuffle(t *testing.T) {
	query := new(State).Compile(`range(4) | [range(.)] | [shuffle(range(30))|join("")] | unique | join("-")`)
	assertString(t, slices.Collect(query(nil)), `[ 0 01-10 012-021-102-120-201-210]`)

	query = new(State).Compile(`def seq(f): range(4) | [range(.)] | [shuffle(range(30)|f)]; seq(.),seq(5) | unique | length`)
	assertString(t, slices.Collect(query(nil)), `[1 1 2 6 1 1 1 1]`)
}
func BenchmarkShuffle(b *testing.B) {
	lists := []string{
		`[1,2]`,
		`[1,2,3]`,
	}

	for _, list := range lists {
		b.Run(list, func(b *testing.B) {
			query := new(State).Compile(`
				def fac($n): if $n==0 then 1 else $n*fac($n-1) end;
				. as $n | ` + constString(list) + `
				| reduce shuffle(range($n*fac(length))) as $perm (
					{};
					.[$perm|join("")] |= .+1
				)
			`)

			ss := slices.Collect(query(b.N))
			assertEqual(b, len(ss), 1)

			buckets := ss[0].(map[string]any)
			for k, v := range buckets {
				v := float64(v.(int))
				b.ReportMetric(v, k)
				v /= float64(b.N)
				v -= 1
				v *= 100
				b.ReportMetric(math.Abs(v), "%-err-"+k)
			}
		})
	}
}
