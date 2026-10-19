package skill

import (
	"strings"
	"testing"
	"testing/fstest"
)

func builtinFS() fstest.MapFS {
	return fstest.MapFS{
		"builtin/beta/SKILL.md":  {Data: []byte("---\nname: beta\ndescription: \"Second. Use when: asked.\"\n---\nBeta body.\n")},
		"builtin/alpha/SKILL.md": {Data: []byte("---\nname: alpha\ndescription: \"First. Use when: asked.\"\n---\nAlpha body.\n")},
	}
}

// Built-ins come from whatever filesystem the app passes: listed in name
// order, their bodies read from it, and none at all from a nil one.
func TestBuiltinsLoadFromAnyFS(t *testing.T) {
	got := Builtins(builtinFS())
	if len(got) != 2 || got[0].Name != "alpha" || got[1].Name != "beta" {
		t.Fatalf("builtins = %+v", got)
	}
	if got[0].Scope != ScopeBuiltin || got[0].Location != BuiltinLocation {
		t.Fatalf("scope %q location %q", got[0].Scope, got[0].Location)
	}
	body, err := got[1].Body()
	if err != nil || !strings.Contains(body, "Beta body.") {
		t.Fatalf("body = %q (%v)", body, err)
	}
	if n := len(Builtins(nil)); n != 0 {
		t.Fatalf("a nil fs carried %d built-ins", n)
	}
}

// Load lists the built-ins after the file skills, and a file skill of the
// same name shadows the built-in with a diagnostic.
func TestLoadShadowsABuiltinWithAFileSkill(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha", "name: alpha\ndescription: \"Mine. Use when: asked.\"", "File body.\n")
	c := Load([]Root{{Path: root, Scope: ScopeProject}}, builtinFS())
	if c.Len() != 2 {
		t.Fatalf("skills = %v", c.Names())
	}
	a, _ := c.Find("alpha")
	if a.Scope != ScopeProject {
		t.Fatalf("alpha scope = %q", a.Scope)
	}
	b, ok := c.Find("beta")
	if !ok || b.Scope != ScopeBuiltin {
		t.Fatalf("beta = %+v", b)
	}
	if len(c.Diagnostics) != 1 || !strings.Contains(c.Diagnostics[0], "shadowed") {
		t.Fatalf("diagnostics = %v", c.Diagnostics)
	}
}
