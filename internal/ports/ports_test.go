package ports

import "testing"

const sample = `p12345
cnode
Lyunis
f23
n*:3000
f24
n*:3000
p999
cpostgres
Lpg
f5
n127.0.0.1:5432
p42
cWeb Content Helper
Lroot
f3
n[::1]:22
p7
cbroken
Lx
f1
nno-port-here
`

func TestParse(t *testing.T) {
	got := Parse(sample)
	want := []Port{
		{PID: 42, Process: "Web Content Helper", User: "root", Addr: "[::1]", Port: 22},
		{PID: 12345, Process: "node", User: "yunis", Addr: "*", Port: 3000},
		{PID: 999, Process: "postgres", User: "pg", Addr: "127.0.0.1", Port: 5432},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d ports, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseEmpty(t *testing.T) {
	if got := Parse(""); len(got) != 0 {
		t.Errorf("expected no ports, got %+v", got)
	}
}
