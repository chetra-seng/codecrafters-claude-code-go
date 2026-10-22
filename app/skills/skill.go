// Package skills offers helpers and type for skills related operations
package skills

type SkillMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Skill is a fully loaded skill: its metadata plus the instructions in its body
type Skill struct {
	SkillMeta
	Body string
}
