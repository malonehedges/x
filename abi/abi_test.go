package abi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/indexsupply/x/abi/schema"
	"github.com/indexsupply/x/tc"
	"github.com/kr/pretty"
	"kr.dev/diff"
)

func TestPad(t *testing.T) {
	cases := []struct {
		desc  string
		input []byte
		want  []byte
	}{
		{
			desc:  "< 4",
			input: []byte{0x2a},
			want:  []byte{0x2a, 0x00, 0x00, 0x00},
		},
		{
			desc:  "= 4",
			input: []byte{0x2a, 0x00, 0x00, 0x00},
			want:  []byte{0x2a, 0x00, 0x00, 0x00},
		},
		{
			desc:  "> 4",
			input: []byte{0x2a, 0x00, 0x00, 0x00, 0x00},
			want:  []byte{0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}
	for _, c := range cases {
		got := rpad(4, c.input)
		if !bytes.Equal(c.want, got) {
			t.Errorf("want: %x got: %x", c.want, got)
		}
	}
}

func TestSolidityVectors(t *testing.T) {
	cases := []struct {
		desc  string
		input *Item
		want  string
	}{
		{
			desc: "https://docs.soliditylang.org/en/latest/abi-spec.html#examples",
			input: Tuple(
				String("dave"),
				Bool(true),
				Array(Uint64(1), Uint64(2), Uint64(3)),
			),
			want: `0000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003`,
		},
		{
			desc: "https://docs.soliditylang.org/en/latest/abi-spec.html#use-of-dynamic-types",
			input: Tuple(
				Array(Array(Uint64(1), Uint64(2)), Array(Uint64(3))),
				Array(String("one"), String("two"), String("three")),
			),
			want: `000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000036f6e650000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000374776f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000057468726565000000000000000000000000000000000000000000000000000000`,
		},
	}
	for _, c := range cases {
		want, err := hex.DecodeString(c.want)
		tc.NoErr(t, err)
		got := Encode(c.input)
		if !bytes.Equal(want, got) {
			t.Errorf("want: %x got: %x", want, got)
		}
	}
}

func TestEncode(t *testing.T) {
	hb := func(s string) []byte {
		s = strings.Map(func(r rune) rune {
			switch {
			case r >= '0' && r <= '9':
				return r
			case r >= 'a' && r <= 'f':
				return r
			default:
				return -1
			}
		}, strings.ToLower(s))
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	cases := []struct {
		desc  string
		input *Item
		want  []byte
	}{
		{
			desc:  "single static",
			input: Uint8(42),
			want:  hb("000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			desc:  "single dynamic",
			input: String("hello world"),
			want: hb(`
				000000000000000000000000000000000000000000000000000000000000000b
				68656c6c6f20776f726c64000000000000000000000000000000000000000000
			`),
		},
		{
			desc:  "dynamic list of static types",
			input: Array(Uint8(42)),
			want: hb(`
				0000000000000000000000000000000000000000000000000000000000000001
				000000000000000000000000000000000000000000000000000000000000002a
			`),
		},
		{
			desc:  "dynamic list of dynamic types",
			input: Array(String("hello"), String("world")),
			want: hb(`
				0000000000000000000000000000000000000000000000000000000000000002
				0000000000000000000000000000000000000000000000000000000000000040
				0000000000000000000000000000000000000000000000000000000000000080
				0000000000000000000000000000000000000000000000000000000000000005
				68656c6c6f000000000000000000000000000000000000000000000000000000
				0000000000000000000000000000000000000000000000000000000000000005
				776f726c64000000000000000000000000000000000000000000000000000000
			`),
		},
		{
			desc:  "static list of static types",
			input: ArrayK(Uint8(42), Uint8(43)),
			want: hb(`
				000000000000000000000000000000000000000000000000000000000000002a
				000000000000000000000000000000000000000000000000000000000000002b
			`),
		},
		{
			desc:  "static list of dynamic types",
			input: ArrayK(String("hello"), String("world")),
			want: hb(`
				0000000000000000000000000000000000000000000000000000000000000040
				0000000000000000000000000000000000000000000000000000000000000080
				0000000000000000000000000000000000000000000000000000000000000005
				68656c6c6f000000000000000000000000000000000000000000000000000000
				0000000000000000000000000000000000000000000000000000000000000005
				776f726c64000000000000000000000000000000000000000000000000000000
			`),
		},
		{
			desc:  "dynamic nested list of dynamic types",
			input: Array(Array(String("hello"), String("world")), Array(String("bye"))),
			want: hb(`
				0000000000000000000000000000000000000000000000000000000000000002
				0000000000000000000000000000000000000000000000000000000000000040
				0000000000000000000000000000000000000000000000000000000000000120
				0000000000000000000000000000000000000000000000000000000000000002
				0000000000000000000000000000000000000000000000000000000000000040
				0000000000000000000000000000000000000000000000000000000000000080
				0000000000000000000000000000000000000000000000000000000000000005
				68656c6c6f000000000000000000000000000000000000000000000000000000
				0000000000000000000000000000000000000000000000000000000000000005
				776f726c64000000000000000000000000000000000000000000000000000000
				0000000000000000000000000000000000000000000000000000000000000001
				0000000000000000000000000000000000000000000000000000000000000020
				0000000000000000000000000000000000000000000000000000000000000003
				6279650000000000000000000000000000000000000000000000000000000000
			`),
		},
		{
			desc:  "tuple with static fields",
			input: Tuple(Uint8(42), Uint8(43)),
			want: hb(`
				000000000000000000000000000000000000000000000000000000000000002a
				000000000000000000000000000000000000000000000000000000000000002b
			`),
		},
		{
			desc:  "tuple with dynamic fields",
			input: Tuple(Uint8(42), Uint8(43), String("hello")),
			want: hb(`
				000000000000000000000000000000000000000000000000000000000000002a
				000000000000000000000000000000000000000000000000000000000000002b
				0000000000000000000000000000000000000000000000000000000000000060
				0000000000000000000000000000000000000000000000000000000000000005
				68656c6c6f000000000000000000000000000000000000000000000000000000
			`),
		},
		{
			desc: "tuple with list of static tuples",
			input: Tuple(
				Array(
					Tuple(Uint8(44), Uint8(45)),
				),
				Array(
					Tuple(Uint8(46), Uint8(47)),
					Tuple(Uint8(48), Uint8(49)),
				),
			),
			want: hb(`
				0000000000000000000000000000000000000000000000000000000000000040
				00000000000000000000000000000000000000000000000000000000000000a0
				0000000000000000000000000000000000000000000000000000000000000001
				000000000000000000000000000000000000000000000000000000000000002c
				000000000000000000000000000000000000000000000000000000000000002d
				0000000000000000000000000000000000000000000000000000000000000002
				000000000000000000000000000000000000000000000000000000000000002e
				000000000000000000000000000000000000000000000000000000000000002f
				0000000000000000000000000000000000000000000000000000000000000030
				0000000000000000000000000000000000000000000000000000000000000031
			`),
		},
		{
			desc: "tuple with list of dynamic tuples",
			input: Tuple(
				Uint8(42),
				Uint8(43),
				String("hello"),
				Array(
					Tuple(
						Uint8(42),
						Uint8(43),
						String("hello"),
					),
				),
			),
			want: hb(`
				000000000000000000000000000000000000000000000000000000000000002a
				000000000000000000000000000000000000000000000000000000000000002b
				0000000000000000000000000000000000000000000000000000000000000080
				00000000000000000000000000000000000000000000000000000000000000c0
				0000000000000000000000000000000000000000000000000000000000000005
				68656c6c6f000000000000000000000000000000000000000000000000000000
				0000000000000000000000000000000000000000000000000000000000000001
				0000000000000000000000000000000000000000000000000000000000000020
				000000000000000000000000000000000000000000000000000000000000002a
				000000000000000000000000000000000000000000000000000000000000002b
				0000000000000000000000000000000000000000000000000000000000000060
				0000000000000000000000000000000000000000000000000000000000000005
				68656c6c6f000000000000000000000000000000000000000000000000000000
			`),
		},
		{
			desc:  "nested static tuples",
			input: Tuple(Tuple(Uint8(42), Uint8(43), Tuple(Uint8(44), Uint8(45)))),
			want: hb(`
				000000000000000000000000000000000000000000000000000000000000002a
				000000000000000000000000000000000000000000000000000000000000002b
				000000000000000000000000000000000000000000000000000000000000002c
				000000000000000000000000000000000000000000000000000000000000002d
			`),
		},
		{
			desc:  "nested dynamic tuples",
			input: Tuple(Tuple(Uint8(42), Tuple(Uint8(43), String("foo")))),
			want: hb(`
				0000000000000000000000000000000000000000000000000000000000000020
				000000000000000000000000000000000000000000000000000000000000002a
				0000000000000000000000000000000000000000000000000000000000000040
				000000000000000000000000000000000000000000000000000000000000002b
				0000000000000000000000000000000000000000000000000000000000000040
				0000000000000000000000000000000000000000000000000000000000000003
				666f6f0000000000000000000000000000000000000000000000000000000000
			`),
		},
	}
	for _, tc := range cases {
		got := Encode(tc.input)
		if !bytes.Equal(got, tc.want) {
			t.Errorf("%q\ngot: %s\nwant: %s\n", tc.desc, dump(got), dump(tc.want))
		}
	}
}

func dump(b []byte) (out string) {
	out += "\n"
	for i := 0; i < len(b); i += 32 {
		out += fmt.Sprintf("%x\n", b[i:i+32])
	}
	return out
}

func TestDecode(t *testing.T) {
	cases := []struct {
		desc  string
		want  *Item
		input schema.Type
	}{
		{
			desc:  "1 static",
			want:  Uint8(0),
			input: schema.Static(),
		},
		{
			desc:  "N static",
			want:  Tuple(Uint64(0), Uint64(1)),
			input: schema.Tuple(schema.Static(), schema.Static()),
		},
		{
			desc:  "1 dynamic",
			want:  String("hello world"),
			input: schema.Dynamic(),
		},
		{
			desc:  "N dynamic",
			want:  Tuple(String("hello"), String("world")),
			input: schema.Tuple(schema.Dynamic(), schema.Dynamic()),
		},
		{
			desc:  "dynamic size list of static types",
			want:  Array(Uint64(0), Uint64(1)),
			input: schema.Array(schema.Static()),
		},
		{
			desc:  "list dynamic",
			want:  Array(String("hello"), String("world")),
			input: schema.Array(schema.Dynamic()),
		},
		{
			desc:  "fixed size list of static types",
			want:  ArrayK(Uint8(42), Uint8(43)),
			input: schema.ArrayK(2, schema.Static()),
		},
		{
			desc:  "fixed size array of dynamic types",
			want:  ArrayK(String("foo"), String("bar")),
			input: schema.ArrayK(2, schema.Dynamic()),
		},
		{
			desc:  "tuple static",
			want:  Tuple(Uint64(0)),
			input: schema.Tuple(schema.Static()),
		},
		{
			desc:  "tuple static and dynamic",
			want:  Tuple(Uint64(0), String("hello")),
			input: schema.Tuple(schema.Static(), schema.Dynamic()),
		},
		{
			desc: "tuple tuple",
			want: Tuple(Uint64(0), Tuple(String("hello"))),
			input: schema.Tuple(
				schema.Static(),
				schema.Tuple(
					schema.Dynamic(),
				),
			),
		},
		{
			desc: "tuple list",
			want: Tuple(Uint64(0), Array(Uint64(1))),
			input: schema.Tuple(
				schema.Static(),
				schema.Array(schema.Static()),
			),
		},
		{
			desc: "tuple with nested static tuple",
			want: Tuple(Uint8(42), Uint8(43), Tuple(Uint8(44), Uint8(45))),
			input: schema.Tuple(
				schema.Static(),
				schema.Static(),
				schema.Tuple(
					schema.Static(),
					schema.Static(),
				),
			),
		},
		{
			desc: "tuple with nested dynamic tuple",
			want: Tuple(Uint8(42), String("foo"), Tuple(Uint8(44), String("bar"))),
			input: schema.Tuple(
				schema.Static(),
				schema.Dynamic(),
				schema.Tuple(
					schema.Static(),
					schema.Dynamic(),
				),
			),
		},
		{
			desc: "tuple with large (> 32) initial field and dynamic second field",
			want: Tuple(
				Tuple(Uint8(42), Uint8(43)),
				Array(String("foo")),
			),
			input: schema.Tuple(
				schema.Tuple(schema.Static(), schema.Static()),
				schema.Array(schema.Dynamic()),
			),
		},
		{
			desc: "tuple with list of tuples",
			want: Tuple(
				Uint8(42),
				Array(
					Tuple(Uint8(43), Uint8(44)),
				),
				Array(
					Tuple(Uint8(45), Uint8(46)),
					Tuple(Uint8(47), Uint8(48)),
				),
			),
			input: schema.Tuple(
				schema.Static(),
				schema.Array(
					schema.Tuple(
						schema.Static(),
						schema.Static(),
					),
				),
				schema.Array(
					schema.Tuple(
						schema.Static(),
						schema.Static(),
					),
				),
			),
		},
	}
	for _, tc := range cases {
		got, _, _ := Decode(debug(tc.desc, t, Encode(tc.want)), tc.input)
		if !got.Equal(tc.want) {
			t.Errorf("decode %q want: %# v got: %# v", tc.desc, pretty.Formatter(tc.want), pretty.Formatter(got))
		}
	}
}

func TestDecode_NumBytes(t *testing.T) {
	cases := []struct {
		item   *Item
		schema schema.Type
		want   int
	}{
		{
			item:   Uint8(42),
			schema: schema.Static(),
			want:   32,
		},
		{
			item:   String("foo"),
			schema: schema.Dynamic(),
			want:   64,
		},
		{
			item:   Bytes([]byte{}),
			schema: schema.Dynamic(),
			want:   32,
		},
		{
			item:   String("foooooooooooooooooooooooooooooooo"), //len=33
			schema: schema.Dynamic(),
			want:   96,
		},
		{
			item:   Tuple(Uint8(42), Uint8(42)),
			schema: schema.Tuple(schema.Static(), schema.Static()),
			want:   64,
		},
		{
			item: Tuple(
				Uint8(42),
				Array(String("foo"), String("bar")),
			),
			schema: schema.Tuple(
				schema.Static(),
				schema.Array(schema.Dynamic()),
			),
			want: 9 * 32,
		},
	}
	for _, tc := range cases {
		_, n, _ := Decode(Encode(tc.item), tc.schema)
		diff.Test(t, t.Errorf, n, tc.want)
	}
}

func TestDecode_ExtraInput(t *testing.T) {
	want := []byte("extra")
	data := Encode(Tuple(Uint8(42), Array(String("foo"), String("bar"))))
	data = append(data, want...)
	_, n, _ := Decode(data, schema.Tuple(
		schema.Static(),
		schema.Array(schema.Dynamic()),
	))
	diff.Test(t, t.Errorf, data[n:], want)
}

func debug(desc string, t *testing.T, b []byte) []byte {
	t.Helper()
	out := fmt.Sprintf("len: %d\n", len(b))
	for i := 0; i < len(b); i += 32 {
		out += fmt.Sprintf("%x\n", b[i:i+32])
	}
	t.Logf("%s\n%s\n", desc, out)
	return b
}

func BenchmarkDecode(b *testing.B) {
	var (
		input  = Encode(Tuple(String("foo"), Array(String("bar"), String("baz"))))
		schema = schema.Tuple(schema.Dynamic(), schema.Array(schema.Dynamic()))
	)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		item, _, _ := Decode(input, schema)
		item.Done()
	}
}

func TestDone(t *testing.T) {
	var i *Item
	i.Done()
}

func TestPartialRead(t *testing.T) {
	var items = make([]*Item, 100)
	for i := uint8(0); i < 100; i++ {
		items[i] = Uint8(i)
	}
	d := Encode(Array(items...))
	item, n, err := Decode(d, schema.ArrayK(1, schema.Static()))
	if err != nil {
		t.Fatalf("got: %s want: nil", err)
	}
	if n != 32 {
		t.Fatalf("got: %d want: 32", n)
	}
	if len(item.l) != 1 {
		t.Error("expected returned item to have 1 element")
	}
}
