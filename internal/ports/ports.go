package ports

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

type Port struct {
	PID     int
	Process string
	User    string
	Addr    string
	Port    int
}

var ErrLsofMissing = errors.New("lsof not found in PATH; install it (e.g. dnf install lsof / apt install lsof)")

func List() ([]Port, error) {
	if _, err := exec.LookPath("lsof"); err != nil {
		return nil, ErrLsofMissing
	}
	// -F gives one field per line (p=pid, c=command, L=user, n=name), so
	// command names with spaces or longer than lsof's 9-char column survive.
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n", "-F", "pcLn").Output()
	// lsof exits 1 when nothing matches; treat empty output as no listeners.
	if err != nil && len(out) == 0 {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof: %w", err)
	}
	return Parse(string(out)), nil
}

func Parse(out string) []Port {
	seen := map[string]bool{}
	var result []Port
	var cur Port
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		val := line[1:]
		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(val)
			if err != nil {
				pid = 0
			}
			cur = Port{PID: pid}
		case 'c':
			cur.Process = val
		case 'L':
			cur.User = val
		case 'n':
			idx := strings.LastIndex(val, ":")
			if idx < 0 || cur.PID == 0 {
				continue
			}
			port, err := strconv.Atoi(val[idx+1:])
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%d:%d", cur.PID, port)
			if seen[key] {
				continue
			}
			seen[key] = true
			p := cur
			p.Addr = val[:idx]
			p.Port = port
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Port != result[j].Port {
			return result[i].Port < result[j].Port
		}
		return result[i].PID < result[j].PID
	})
	return result
}

func Kill(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("kill %d: %w", pid, err)
	}
	return nil
}
