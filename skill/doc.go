// Package skill discovers and loads Agent Skills: directories holding a
// SKILL.md whose frontmatter names the skill and says when to use it, and whose
// body is the instructions to follow once it applies. The format is the Agent
// Skills specification (https://agentskills.io/specification), so a skill
// written for any other harness loads here unchanged, and one written here
// loads there.
//
// A catalog is disclosed in three tiers, and the tiers are the point: the name
// and description of every skill go into the system prompt at startup, the body
// of one skill enters the conversation when it is activated, and the files
// beside it are read on demand by the ordinary file tools.
package skill
