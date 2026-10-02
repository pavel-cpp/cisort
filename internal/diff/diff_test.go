package diff

import (
	"strings"
	"testing"
)

func TestUnified(t *testing.T) {
	a := "1\n2\n3\n4\n5\n#include <b>\n#include <a>\n6\n7\n8\n9\n10\n11\n12\n#include <d>\r\n#include <c>"
	b := "1\n2\n3\n4\n5\n#include <a>\n#include <b>\n6\n7\n8\n9\n10\n11\n12\n#include <c>\n#include <d>\r\n"
	want := `--- a/x.cpp
+++ b/x.cpp
@@ -3,8 +3,8 @@
 3
 4
 5
-#include <b>
 #include <a>
+#include <b>
 6
 7
 8
@@ -12,5 +12,5 @@
 10
 11
 12
+#include <c>
 #include <d>
-#include <c>
\ No newline at end of file
`
	if got := Unified("a/x.cpp", "b/x.cpp", []byte(a), []byte(b)); got != want {
		t.Errorf("Unified() =\n%s\nwant\n%s", got, want)
	}
}

func TestUnifiedEqual(t *testing.T) {
	if got := Unified("a", "b", []byte("x\n"), []byte("x\n")); got != "" {
		t.Errorf("Unified() of equal inputs = %q, want empty", got)
	}
}

func TestUnifiedLarge(t *testing.T) {
	a := strings.Repeat("a\n", 3000)
	b := strings.Repeat("b\n", 3000)
	got := Unified("a", "b", []byte(a), []byte(b))
	if !strings.Contains(got, "@@ -1,3000 +1,3000 @@") {
		t.Errorf("Unified() of a large change has an unexpected header:\n%.200s", got)
	}
}
