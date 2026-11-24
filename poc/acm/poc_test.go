// Package acm_test runs the end-to-end PoC: the AIF integrator over a real ACM
// adapter and a throwaway memory. run.sh is the single entry point so the demo
// the person runs and the test are the same code path.
package acm_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestACMPreservesItsGuarantees(t *testing.T) {
	out, err := exec.Command("bash", "run.sh").CombinedOutput()
	t.Logf("poc/acm/run.sh:\n%s", out)
	if err != nil {
		t.Fatalf("run.sh falló: %v", err)
	}
	text := string(out)
	for _, scenario := range []string{"S1", "S2", "S3", "S4"} {
		want := "VEREDICT " + scenario + ": PASS"
		if !strings.Contains(text, want) {
			t.Errorf("falta %q", want)
		}
	}
	if strings.Contains(text, ": FAIL") {
		t.Errorf("algún escenario falló; salida:\n%s", text)
	}
	if !strings.Contains(text, "RESULT: 4/4 PASS") {
		t.Error("no se informó de 4/4 PASS")
	}
}
