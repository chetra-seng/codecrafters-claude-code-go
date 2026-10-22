package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConstructSkillPrompt construct agent prompt telling the skills available based on skill metadata
func ConstructSkillPrompt(skills []SkillMeta) string {
	var sb strings.Builder
	sb.WriteString("You have access to the following skills:\n\n")

	for _, skill := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", skill.Name, skill.Description))
	}

	return sb.String()
}

// ParseInvocation reports whether prompt invokes a skill ("/name args...")
// and returns the skill name and the remaining arguments
func ParseInvocation(prompt string) (name string, args string, ok bool) {
	if !strings.HasPrefix(prompt, "/") {
		return "", "", false
	}

	fields := strings.Fields(prompt[1:])
	if len(fields) == 0 {
		return "", "", false
	}

	name = fields[0]
	args = strings.TrimSpace(strings.TrimPrefix(prompt[1:], name))

	return name, args, true
}

// LoadSkillMetas return the metadata of every skill found under the given path
func LoadSkillMetas(path string) []SkillMeta {
	metas := make([]SkillMeta, 0)

	entries, err := os.ReadDir(path)
	if err != nil {
		return metas
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		s, err := readSkill(filepath.Join(path, entry.Name()))
		if err != nil {
			continue
		}

		metas = append(metas, s.SkillMeta)
	}

	return metas
}

// LoadSkill return the skill with the given name under path, including its body
func LoadSkill(path string, name string) (Skill, error) {
	// Keep the lookup inside path, e.g. reject "/../../etc"
	if name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return Skill{}, fmt.Errorf("invalid skill name %q", name)
	}

	return readSkill(filepath.Join(path, name))
}

// readSkill parse the SKILL.md inside dir into metadata and body
func readSkill(dir string) (Skill, error) {
	content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return Skill{}, err
	}

	front, body := splitFrontmatter(string(content))

	var s Skill
	if err := yaml.Unmarshal([]byte(front), &s.SkillMeta); err != nil {
		return Skill{}, fmt.Errorf("parse frontmatter of %s: %w", dir, err)
	}

	if s.Name == "" {
		s.Name = filepath.Base(dir)
	}
	s.Body = strings.TrimSpace(body)

	return s, nil
}

// splitFrontmatter split a markdown document into its "---" delimited
// frontmatter and the body that follows. Without frontmatter, everything is body
func splitFrontmatter(content string) (front string, body string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	if strings.TrimSpace(lines[0]) != "---" {
		return "", content
	}

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n")
		}
	}

	return "", content
}
