package skill_test

import (
	"fmt"
	"testing/fstest"

	"github.com/rfizzle/agentcore/skill"
)

// Load finds the skills under the roots an app names and adds the built-ins
// carried in an fs.FS the app passes, usually an embed. The fs.FS keeps the
// module free of any one app's files: the built-ins are read from
// builtin/<name>/SKILL.md in whatever filesystem is given.
func ExampleLoad() {
	builtins := fstest.MapFS{
		"builtin/release/SKILL.md": {Data: []byte("---\n" +
			"name: release\n" +
			"description: \"Cut a release. Use when: asked to ship.\"\n" +
			"---\n" +
			"Tag the commit, then publish.\n")},
	}

	c := skill.Load(nil, builtins)
	fmt.Println(c.Names())

	s, _ := c.Find("release")
	body, _ := s.Body()
	fmt.Printf("%s (%s): %s\n", s.Name, s.Location, body)
	fmt.Println(skill.PromptBlock(c) != "")
	// Output:
	// [release]
	// release (built-in): Tag the commit, then publish.
	// true
}
